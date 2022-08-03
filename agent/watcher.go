package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"github.com/thomasglachant/spooter/core"
)

const (
	watcherLogPrefix = "watcher"
	watcherTimer     = 1 * time.Second

	eventNameLogDiscover = "agent.log.discover"

	parserModeRegex = "regex"
	parserModeJSON  = "json"
)

type LogDiscoverEvent struct {
	Log *core.LogLine
}

func (event *LogDiscoverEvent) Name() string {
	return eventNameLogDiscover
}

func (event *LogDiscoverEvent) Data() interface{} {
	return event.Log
}

type currentWatching struct {
	parser   *ParserConfigStruct
	fileName string
	tail     *tail.Tail
}

type Watcher struct {
	mu       sync.Mutex
	exitChan chan bool

	currentTails map[string]*currentWatching
	regexCache   map[string]*regexp.Regexp
}

func (watcher *Watcher) Run() error {
	watcher.regexCache = make(map[string]*regexp.Regexp)
	watcher.currentTails = map[string]*currentWatching{}
	watcher.exitChan = make(chan bool)

	core.ProcessInfiniteLoop(watcherTimer, watcher.exitChan, func() {
		// execute watcher
		if err := watcher.discoverFilesToWatch(); err != nil {
			core.Logger.Errorf(watcherLogPrefix, "Error while discover files to watch: %s", err)
		}
		watcher.cleanUpVanishedFiles()
	})

	return nil
}

func (watcher *Watcher) HandleStop() {
	for k := range watcher.currentTails {
		watcher.endWatchFromTailKey(k)
	}
	watcher.exitChan <- true
}

func (watcher *Watcher) cleanUpVanishedFiles() {
	for k, v := range watcher.currentTails {
		if _, err := os.Stat(v.fileName); err != nil {
			watcher.endWatchFromTailKey(k)
		}
	}
}

func (watcher *Watcher) discoverFilesToWatch() error {
	for _, parser := range config.Parsers {
		includedFiles, err := core.GetFilesMatchingPatterns(parser.FilesIncluded)
		if err != nil {
			return fmt.Errorf("error while retrieve included files %s: %s", parser.FilesIncluded, err)
		}
		excludedFiles, err := core.GetFilesMatchingPatterns(parser.FilesExcluded)
		if err != nil {
			return fmt.Errorf("error while retrieve excluded files %s: %s", parser.FilesIncluded, err)
		}

		for _, file := range includedFiles {
			// ignore if file excluded
			if core.SliceContains(excludedFiles, file) {
				continue
			}
			// ignore if file already watched
			if _, ok := watcher.currentTails[watcher.genTailKey(parser, file)]; ok {
				continue
			}
			// start watching file
			go watcher.startWatchFile(parser, file)
		}
	}

	return nil
}

func (watcher *Watcher) genTailKey(parser *ParserConfigStruct, file string) string {
	return fmt.Sprintf("%s-%s", sha256.New().Sum([]byte(parser.Name)), file)
}

func (watcher *Watcher) startWatchFile(parser *ParserConfigStruct, file string) {
	core.Logger.Infof(watcherLogPrefix, "Start watching file %s", file)

	t, err := tail.TailFile(
		file,
		tail.Config{
			Follow: true,
			ReOpen: true,
			Location: &tail.SeekInfo{
				Offset: 0,
				Whence: io.SeekEnd,
			},
			Logger: tail.DiscardingLogger,
		})
	if err != nil {
		panic(err)
	}

	cur := &currentWatching{
		parser:   parser,
		fileName: file,
		tail:     t,
	}

	watcher.mu.Lock()
	watcher.currentTails[watcher.genTailKey(parser, file)] = cur
	watcher.mu.Unlock()

	for line := range t.Lines {
		core.Logger.Debugf(watcherLogPrefix, "Receive Line: %s", line.Text)

		var log *core.LogLine
		var err error
		log, err = watcher.handleLine(cur, line.Text)
		if err != nil {
			core.Logger.Errorf(watcherLogPrefix, "Error while handle line: %s", err)
			continue
		}

		core.Logger.Debugf(watcherLogPrefix, "Line handled")
		for k, v := range log.Fields {
			core.Logger.Debugf(watcherLogPrefix, "Field %s: %s", k, v)
		}

		core.EventDispatcher.Dispatch(&LogDiscoverEvent{Log: log})
	}

	watcher.mu.Lock()
	delete(watcher.currentTails, watcher.genTailKey(parser, file))
	watcher.mu.Unlock()
}

func (watcher *Watcher) endWatchFromTailKey(tailKey string) {
	if _, ok := watcher.currentTails[tailKey]; !ok {
		return
	}

	core.Logger.Infof(watcherLogPrefix, "End watching file %s", watcher.currentTails[tailKey].fileName)

	watcher.currentTails[tailKey].tail.Cleanup()
	_ = watcher.currentTails[tailKey].tail.Stop()

	watcher.mu.Lock()
	delete(watcher.currentTails, tailKey)
	watcher.mu.Unlock()
}

//nolint:gocyclo
func (watcher *Watcher) handleLine(fileWatcher *currentWatching, line string) (*core.LogLine, error) {
	log := &core.LogLine{
		Metadata: core.LogMetadata{
			AgentVersion: version,
			Application:  config.Application,
			Server:       config.Server,
			Filename:     fileWatcher.fileName,
			Parser:       fileWatcher.parser.Name,
			CaptureDate:  time.Now(),
		},
		Date:   time.Now(),
		Raw:    line,
		Fields: map[string]string{},
	}

	switch {
	case fileWatcher.parser.Mode == parserModeRegex:
		var regex *regexp.Regexp
		var ok bool
		if regex, ok = watcher.regexCache[fileWatcher.parser.Name]; !ok {
			watcher.mu.Lock()
			watcher.regexCache[fileWatcher.parser.Name] = regexp.MustCompile(fileWatcher.parser.RegexPattern)
			regex = watcher.regexCache[fileWatcher.parser.Name]
			watcher.mu.Unlock()
		}
		matches := regex.FindStringSubmatch(line)
		for i, name := range regex.SubexpNames() {
			if i != 0 && name != "" {
				log.Fields[name] = matches[i]
			}
		}
	case fileWatcher.parser.Mode == parserModeJSON:
		jsonData := map[string]interface{}{}
		if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
			return nil, fmt.Errorf("unable to parse line as json: %s", err)
		}
		for internalFieldName, jsonField := range fileWatcher.parser.JSONFields {
			// json field contain "." => use json path
			if strings.Contains(jsonField, ".") {
				splitByDot := strings.Split(jsonField, ".")
				var cur interface{} = jsonData
				if len(splitByDot) > 1 {
					for _, part := range splitByDot {
						if _, ok := cur.(map[string]interface{})[part]; ok {
							cur = cur.(map[string]interface{})[part]
						} else {
							cur = nil
							break
						}
					}
					jsonData[jsonField] = cur
				}
			}

			// jsonField not exists
			if _, ok := jsonData[jsonField]; !ok {
				continue
			}
			// jsonField is nil
			if jsonData[jsonField] == nil {
				log.Fields[internalFieldName] = ""
				continue
			}

			switch reflect.TypeOf(jsonData[jsonField]).Kind() {
			case reflect.String:
				log.Fields[internalFieldName] = jsonData[jsonField].(string)
			case reflect.Float64:
				if core.IsDecimal(jsonData[jsonField].(float64)) {
					log.Fields[internalFieldName] = fmt.Sprintf("%d", int64(jsonData[jsonField].(float64)))
				} else {
					log.Fields[internalFieldName] = fmt.Sprintf("%f", jsonData[jsonField])
				}
			case reflect.Int:
				log.Fields[internalFieldName] = fmt.Sprintf("%d", jsonData[jsonField].(int64))
			case reflect.Map:
				content, _ := json.Marshal(jsonData[jsonField])
				log.Fields[internalFieldName] = string(content)
			default:
				log.Fields[internalFieldName] = fmt.Sprintf("%v", jsonData[jsonField])
			}
		}
	default:
		return nil, fmt.Errorf("unknown mode %s", fileWatcher.parser.Mode)
	}

	return log, nil
}
