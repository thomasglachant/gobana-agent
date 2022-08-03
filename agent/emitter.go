package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/thomasglachant/spooter/core"
)

const (
	emitterLogPrefix              = "emitter"
	emitterTimeout                = 30 * time.Second
	emitterFrequency              = 5 * time.Second
	emitterMaxLogToKeepInBuffer   = 200000
	emitterMaxLogToSendAtSameTime = 10000

	emitterEndpoint = "http://%s:%d/v1/logs"
)

type Emitter struct {
	pendingMutex sync.Mutex
	isFlushing   sync.Mutex
	exitChan     chan bool
	logBuffer    []*core.LogLine

	client     *http.Client
	postLogURL string
}

func (emitter *Emitter) Run() error {
	emitter.exitChan = make(chan bool)

	// subscribe events
	subscriptionID := core.EventDispatcher.Subscribe(core.EventDescription{
		Name:     eventNameLogDiscover,
		Priority: 0,
		Callback: func(data interface{}) {
			emitter.addLogLine(data.(*core.LogLine))
		},
	})
	defer core.EventDispatcher.Unsubscribe(subscriptionID)

	// configure
	emitter.client = &http.Client{Timeout: emitterTimeout, Transport: &http.Transport{
		DisableKeepAlives: true,
	}}
	emitter.postLogURL = fmt.Sprintf(emitterEndpoint, config.Emitter.Server, config.Emitter.Port)

	// run
	core.ProcessInfiniteLoop(emitterFrequency, emitter.exitChan, func() {
		if len(emitter.logBuffer) > 0 {
			core.Logger.Infof(emitterLogPrefix, "%d logs pending", len(emitter.logBuffer))
		}
		// remove older data
		emitter.cleanupBuffer(true)

		// flush pending logs
		emitter.flush()
	})
	// execute last flush before exiting
	emitter.flush()

	return nil
}

func (emitter *Emitter) HandleStop() {
	emitter.exitChan <- true
}

func (emitter *Emitter) addLogLine(logLine *core.LogLine) {
	go func() {
		core.Logger.Debugf(emitterLogPrefix, "Add log to pool")

		emitter.pendingMutex.Lock()
		defer emitter.pendingMutex.Unlock()

		emitter.logBuffer = append(emitter.logBuffer, logLine)
		emitter.cleanupBuffer(false)
	}()
}

func (emitter *Emitter) cleanupBuffer(withMutexlock bool) {
	if int64(len(emitter.logBuffer)) <= emitterMaxLogToKeepInBuffer {
		return
	}

	if withMutexlock {
		emitter.pendingMutex.Lock()
	}
	startL := int64(len(emitter.logBuffer)) - emitterMaxLogToKeepInBuffer
	emitter.logBuffer = emitter.logBuffer[startL:]
	if withMutexlock {
		emitter.pendingMutex.Unlock()
	}
}

func (emitter *Emitter) flush() {
	if len(emitter.logBuffer) == 0 {
		return
	}

	core.Logger.Debugf(emitterLogPrefix, "Flush %d pending logs", len(emitter.logBuffer))

	// lock resources
	emitter.isFlushing.Lock()
	defer emitter.isFlushing.Unlock()

	// get logs from queue
	emitter.pendingMutex.Lock()

	var nbLogToEmit int
	if len(emitter.logBuffer) > emitterMaxLogToSendAtSameTime {
		nbLogToEmit = emitterMaxLogToSendAtSameTime
	} else {
		nbLogToEmit = len(emitter.logBuffer)
	}
	logLines := emitter.logBuffer[:nbLogToEmit]
	emitter.logBuffer = emitter.logBuffer[nbLogToEmit:]
	emitter.pendingMutex.Unlock()

	//
	// compress and encrypt message
	var encryptedData []byte
	var encryptError error
	encryptedData, encryptError = core.EncryptMessage(&core.SynchronizeLogsMessage{Logs: logLines}, config.Emitter.Secret)
	if encryptError != nil {
		core.Logger.Errorf(emitterLogPrefix, "Error encrypting data: %s", encryptError)
		return
	}

	// create request
	req, _ := http.NewRequest(
		http.MethodPost,
		emitter.postLogURL,
		bytes.NewBuffer(encryptedData),
	)
	req.SetBasicAuth(core.SyncLogin, config.Emitter.Secret)

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
		emitter.logBuffer = append(logLines, emitter.logBuffer...)
		emitter.cleanupBuffer(false)
		emitter.pendingMutex.Unlock()
	}

	core.Logger.Debugf(emitterLogPrefix, "%d logs emitted", len(logLines))
}
