package agent

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

	"gobana-agent/core"
)

const (
	watcherLogPrefix = "watcher"
	watcherTimer     = 1 * time.Second

	eventNameEntryDiscover = "agent.log.discover"

	parserModeRegex = "regex"
	parserModeJSON  = "json"
)

type EntryDiscoverEvent struct {
	Entry *core.Entry
}

func (event *EntryDiscoverEvent) Name() string {
	return eventNameEntryDiscover
}

func (event *EntryDiscoverEvent) Data() interface{} {
	return event.Entry
}

type currentWatching struct {
	parser   *ParserConfigStruct
	fileName string
	tail     *tail.Tail
}

type WatcherProcess struct {
	mu       sync.Mutex
	exitChan chan bool

	currentTails map[string]*currentWatching
	regexCache   map[string]*regexp.Regexp
}

func (watcher *WatcherProcess) Name() string {
	return watcherLogPrefix
}

func (watcher *WatcherProcess) Run() error {
	watcher.regexCache = make(map[string]*regexp.Regexp)
	watcher.currentTails = map[string]*currentWatching{}
	watcher.exitChan = make(chan bool)

	core.ProcessInfiniteLoop(watcherTimer, watcher.exitChan, func() {
		// execute Watcher
		if err := watcher.discoverFilesToWatch(); err != nil {
			core.Logger.Errorf(watcherLogPrefix, "Error while discover files to watch: %s", err)
		}
		watcher.cleanUpVanishedFiles()
	})

	return nil
}

func (watcher *WatcherProcess) HandleStop() {
	for k := range watcher.currentTails {
		watcher.endWatchFromTailKey(k)
	}
	watcher.exitChan <- true
}

func (watcher *WatcherProcess) cleanUpVanishedFiles() {
	for k, v := range watcher.currentTails {
		if _, err := os.Stat(v.fileName); err != nil {
			watcher.endWatchFromTailKey(k)
		}
	}
}

func (watcher *WatcherProcess) discoverFilesToWatch() error {
	for _, parser := range AppConfig.Parsers {
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

func (watcher *WatcherProcess) genTailKey(parser *ParserConfigStruct, file string) string {
	return fmt.Sprintf("%s-%s", sha256.New().Sum([]byte(parser.Name)), file)
}

func (watcher *WatcherProcess) startWatchFile(parser *ParserConfigStruct, file string) {
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

		var entry *core.Entry
		var err error
		entry, err = watcher.handleLine(cur, line.Text)
		if err != nil {
			core.Logger.Errorf(watcherLogPrefix, "Error while handle line: %s", err)
			continue
		}

		core.Logger.Debugf(watcherLogPrefix, "Line handled")
		for k, v := range entry.Fields {
			core.Logger.Debugf(watcherLogPrefix, "Field %s: %s", k, v)
		}

		core.EventDispatcher.Dispatch(&EntryDiscoverEvent{Entry: entry})
	}

	watcher.mu.Lock()
	delete(watcher.currentTails, watcher.genTailKey(parser, file))
	watcher.mu.Unlock()
}

func (watcher *WatcherProcess) endWatchFromTailKey(tailKey string) {
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

func (watcher *WatcherProcess) handleLine(fileWatcher *currentWatching, line string) (*core.Entry, error) {
	// default values
	entry := &core.Entry{
		Metadata: core.EntryMetadata{
			AgentVersion: AppVersion,
			Application:  AppConfig.Application,
			Workspace:    AppConfig.Emitter.WorkspaceID,
			Server:       AppConfig.Server,
			Filename:     fileWatcher.fileName,
			Parser:       fileWatcher.parser.Name,
			CaptureDate:  time.Now(),
		},
		Date:   time.Now(),
		Raw:    line,
		Fields: map[string]string{},
	}

	// parse log line
	switch {
	case fileWatcher.parser.Mode == parserModeRegex:
		if err := watcher.handleParseRegex(fileWatcher, entry, line); err != nil {
			return nil, fmt.Errorf("error while handle regex: %s", err)
		}
	case fileWatcher.parser.Mode == parserModeJSON:
		if err := watcher.handleParseJSON(fileWatcher, entry, line); err != nil {
			return nil, fmt.Errorf("error while handle json: %s", err)
		}
	default:
		return nil, fmt.Errorf("unknown mode %s", fileWatcher.parser.Mode)
	}

	// extract date from entry
	if err := watcher.extractDate(fileWatcher, entry); err != nil {
		return nil, fmt.Errorf("error while extract date: %s", err)
	}

	return entry, nil
}

func (watcher *WatcherProcess) handleParseRegex(fileWatcher *currentWatching, entry *core.Entry, line string) error {
	var regex *regexp.Regexp
	var ok bool
	if regex, ok = watcher.regexCache[fileWatcher.parser.Name]; !ok {
		watcher.mu.Lock()
		watcher.regexCache[fileWatcher.parser.Name] = regexp.MustCompile(fileWatcher.parser.RegexPattern)
		regex = watcher.regexCache[fileWatcher.parser.Name]
		watcher.mu.Unlock()
	}

	matches := regex.FindStringSubmatch(line)
	// if no match, return error
	if len(matches) == 0 {
		return fmt.Errorf("line not match regex (%s)", line)
	}
	// extract fields from matches
	for i, name := range regex.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		if len(matches) >= i+1 {
			entry.Fields[name] = matches[i]
		} else {
			entry.Fields[name] = ""
		}
	}

	return nil
}

func (watcher *WatcherProcess) handleParseJSON(fileWatcher *currentWatching, entry *core.Entry, line string) error {
	jsonData := map[string]interface{}{}
	if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
		return fmt.Errorf("unable to parse line as json: %s", err)
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
			entry.Fields[internalFieldName] = ""
			continue
		}

		switch reflect.TypeOf(jsonData[jsonField]).Kind() {
		case reflect.String:
			entry.Fields[internalFieldName] = jsonData[jsonField].(string)
		case reflect.Float64:
			if core.IsDecimal(jsonData[jsonField].(float64)) {
				entry.Fields[internalFieldName] = fmt.Sprintf("%d", int64(jsonData[jsonField].(float64)))
			} else {
				entry.Fields[internalFieldName] = fmt.Sprintf("%f", jsonData[jsonField])
			}
		case reflect.Int:
			entry.Fields[internalFieldName] = fmt.Sprintf("%d", jsonData[jsonField].(int64))
		case reflect.Map:
			content, _ := json.Marshal(jsonData[jsonField])
			entry.Fields[internalFieldName] = string(content)
		default:
			entry.Fields[internalFieldName] = fmt.Sprintf("%v", jsonData[jsonField])
		}
	}

	return nil
}

func (watcher *WatcherProcess) extractDate(fileWatcher *currentWatching, entry *core.Entry) error {
	// date extraction
	if fileWatcher.parser.DateExtract.Field != "" {
		// search for date field
		if _, ok := entry.Fields[fileWatcher.parser.DateExtract.Field]; ok {
			// date field found
			date, err := time.Parse(fileWatcher.parser.DateExtract.Format, entry.Fields[fileWatcher.parser.DateExtract.Field])
			if err != nil {
				core.Logger.Errorf(watcherLogPrefix, "Error while parsing date: %s", err)
				return err
			}
			entry.Date = date
		}
	}

	return nil
}
