package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"gobana-agent/core"
)

const (
	emitterLogPrefix                  = "emitter"
	emitterTimeout                    = 30 * time.Second
	emitterFrequency                  = 5 * time.Second
	emitterMaxEntriesToKeepInBuffer   = 200000
	emitterMaxEntriesToSendAtSameTime = 10000

	emitterEndpoint = "http://%s:%d/v1/entries"
)

type EmitterProcess struct {
	pendingMutex  sync.Mutex
	isFlushing    sync.Mutex
	exitChan      chan bool
	entriesBuffer []*core.Entry

	client       *http.Client
	postEntryURL string
}

func (emitter *EmitterProcess) Name() string {
	return emitterLogPrefix
}

func (emitter *EmitterProcess) Run() error {
	emitter.exitChan = make(chan bool)

	// subscribe events
	subscriptionID := core.EventDispatcher.Subscribe(core.EventDescription{
		Name:     eventNameEntryDiscover,
		Priority: 0,
		Callback: func(data interface{}) {
			emitter.addEntry(data.(*core.Entry))
		},
	})
	defer core.EventDispatcher.Unsubscribe(subscriptionID)

	// configure
	emitter.client = &http.Client{Timeout: emitterTimeout, Transport: &http.Transport{
		DisableKeepAlives: true,
	}}
	emitter.postEntryURL = fmt.Sprintf(emitterEndpoint, AppConfig.Emitter.Server, AppConfig.Emitter.Port)

	// run
	core.ProcessInfiniteLoop(emitterFrequency, emitter.exitChan, func() {
		if len(emitter.entriesBuffer) > 0 {
			core.Logger.Infof(emitterLogPrefix, "%d entries pending", len(emitter.entriesBuffer))
		}
		// remove older data
		emitter.cleanupBuffer(true)

		// flush pending entries
		emitter.flush()
	})
	// execute last flush before exiting
	emitter.flush()

	return nil
}

func (emitter *EmitterProcess) HandleStop() {
	emitter.exitChan <- true
}

func (emitter *EmitterProcess) addEntry(entry *core.Entry) {
	go func() {
		core.Logger.Debugf(emitterLogPrefix, "Add log to pool")

		emitter.pendingMutex.Lock()
		defer emitter.pendingMutex.Unlock()

		emitter.entriesBuffer = append(emitter.entriesBuffer, entry)
		emitter.cleanupBuffer(false)
	}()
}

func (emitter *EmitterProcess) cleanupBuffer(withMutexlock bool) {
	if int64(len(emitter.entriesBuffer)) <= emitterMaxEntriesToKeepInBuffer {
		return
	}

	if withMutexlock {
		emitter.pendingMutex.Lock()
	}
	startL := int64(len(emitter.entriesBuffer)) - emitterMaxEntriesToKeepInBuffer
	emitter.entriesBuffer = emitter.entriesBuffer[startL:]
	if withMutexlock {
		emitter.pendingMutex.Unlock()
	}
}

func (emitter *EmitterProcess) flush() {
	if len(emitter.entriesBuffer) == 0 {
		return
	}

	core.Logger.Debugf(emitterLogPrefix, "Flush %d pending entries", len(emitter.entriesBuffer))

	// lock resources
	emitter.isFlushing.Lock()
	defer emitter.isFlushing.Unlock()

	// get entries from queue
	emitter.pendingMutex.Lock()

	var nbToEmit int
	if len(emitter.entriesBuffer) > emitterMaxEntriesToSendAtSameTime {
		nbToEmit = emitterMaxEntriesToSendAtSameTime
	} else {
		nbToEmit = len(emitter.entriesBuffer)
	}
	entries := emitter.entriesBuffer[:nbToEmit]
	emitter.entriesBuffer = emitter.entriesBuffer[nbToEmit:]
	emitter.pendingMutex.Unlock()

	//
	// compress and encrypt message
	var encryptedData []byte
	var encryptError error
	encryptedData, encryptError = core.EncryptMessage(&core.SynchronizeEntriesMessage{Entries: entries}, AppConfig.Emitter.Secret)
	if encryptError != nil {
		core.Logger.Errorf(emitterLogPrefix, "Error encrypting data: %s", encryptError)
		return
	}

	// create request
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		emitter.postEntryURL,
		bytes.NewBuffer(encryptedData),
	)
	req.SetBasicAuth(core.SyncLogin, AppConfig.Emitter.Secret)

	// do request
	resp, err := emitter.client.Do(req)

	//  error during request
	isSuccess := true
	if err != nil {
		isSuccess = false
		core.Logger.Errorf(emitterLogPrefix, "error during http request %v", err)
	}

	// response not successful
	if err == nil && resp.StatusCode != http.StatusOK {
		isSuccess = false

		var responseBody []byte
		var errBody error
		responseBody, errBody = io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if errBody != nil {
			core.Logger.Errorf(emitterLogPrefix, "Unable to read response body %v", errBody)
		}

		core.Logger.Errorf(emitterLogPrefix, "Unsuccessfully http request %v (%s)", resp.StatusCode, string(responseBody))
	}

	// not success: keep logs for next try
	if !isSuccess {
		emitter.pendingMutex.Lock()
		emitter.entriesBuffer = append(entries, emitter.entriesBuffer...)
		emitter.cleanupBuffer(false)
		emitter.pendingMutex.Unlock()
	}

	core.Logger.Debugf(emitterLogPrefix, "%d entries emitted", len(entries))
}
