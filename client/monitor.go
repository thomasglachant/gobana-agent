package client

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"spotter/core"
)

const (
	monitorTimer     = 1 * time.Second
	monitorDelay     = 100 * time.Millisecond
	monitorLogPrefix = "monitor"
)

type currentMonitoring struct {
	lookupName string
	fileName   string
	tail       *tail.Tail
}

type Monitor struct {
	mu            sync.Mutex
	askForStop    bool
	lastExecution time.Time

	currentTails map[string]currentMonitoring
}

func (process *Monitor) Run() error {
	process.currentTails = map[string]currentMonitoring{}

	process.askForStop = false
	for !process.askForStop {
		if time.Now().Before(process.lastExecution.Add(monitorTimer)) {
			time.Sleep(monitorDelay)

			continue
		}
		process.lastExecution = time.Now()

		process.discoverFilesToMonitor()
		process.cleanUpVanishedFiles()
	}

	return nil
}

func (process *Monitor) HandleStop() {
	for k := range process.currentTails {
		process.endMonitorFromTailKey(k)
	}
	process.askForStop = true
}

func (process *Monitor) genTailKey(lookup core.LookupConfig, file string) string {
	return fmt.Sprintf("%s-%s", lookup.Name, file)
}

func (process *Monitor) getMatchingFiles() []struct {
	lookup core.LookupConfig
	files  []string
} {
	var out []struct {
		lookup core.LookupConfig
		files  []string
	}
	for _, lookup := range config.Lookups {
		for _, filePattern := range lookup.Files {
			matches, err := filepath.Glob(filePattern)
			if err != nil {
				core.Logger.Errorf(monitorLogPrefix, "Error while retrieve file list pattern %s: %s", filePattern, err)
			}

			fileMatches := []string{}
			for _, match := range matches {
				if fileInfo, err := os.Stat(match); err == nil {
					if fileInfo.Mode().IsRegular() {
						fileMatches = append(fileMatches, match)
					}
				}
			}

			out = append(out, struct {
				lookup core.LookupConfig
				files  []string
			}{
				lookup: lookup,
				files:  fileMatches,
			})
		}
	}
	return out
}

func (process *Monitor) cleanUpVanishedFiles() {
	for k, v := range process.currentTails {
		if _, err := os.Stat(v.fileName); err != nil {
			process.endMonitorFromTailKey(k)
		}
	}
}

func (process *Monitor) discoverFilesToMonitor() {
	for _, v := range process.getMatchingFiles() {
		for _, file := range v.files {
			tailKey := process.genTailKey(v.lookup, file)
			if _, ok := process.currentTails[tailKey]; !ok {
				go process.startMonitorFile(v.lookup, file)
			}
		}
	}
}

func (process *Monitor) startMonitorFile(lookup core.LookupConfig, file string) {
	// Start monitor file
	core.Logger.Infof(monitorLogPrefix, "Start monitoring file %s", file)

	t, err := tail.TailFile(
		file,
		tail.Config{
			Follow: true,
			ReOpen: true,
			Location: &tail.SeekInfo{
				Offset: 0,
				Whence: 2,
			},
			Logger: tail.DiscardingLogger,
		})
	if err != nil {
		panic(err)
	}

	process.mu.Lock()
	process.currentTails[process.genTailKey(lookup, file)] = currentMonitoring{
		lookupName: lookup.Name,
		fileName:   file,
		tail:       t,
	}
	process.mu.Unlock()

	patterns := map[string]*regexp.Regexp{}
	for _, pattern := range lookup.Patterns {
		if pattern.Type == "regex" {
			patterns[pattern.Name] = regexp.MustCompile(pattern.Value)
		}
	}

	for line := range t.Lines {
		concernedPatterns := []string{}
		for patternName, patternReg := range patterns {
			if patternReg.MatchString(line.Text) {
				concernedPatterns = append(concernedPatterns, patternName)
			}
		}

		if len(concernedPatterns) > 0 {
			alerter.addAlert(&Alert{
				Date:     time.Now(),
				Filename: file,
				Patterns: concernedPatterns,
				Content:  line.Text,
			})
		}
	}

	process.mu.Lock()
	delete(process.currentTails, process.genTailKey(lookup, file))
	process.mu.Unlock()
}

func (process *Monitor) endMonitorFromTailKey(tailKey string) {
	if _, ok := process.currentTails[tailKey]; !ok {
		return
	}

	core.Logger.Infof(monitorLogPrefix, "End monitoring file %s", process.currentTails[tailKey].fileName)

	process.currentTails[tailKey].tail.Cleanup()
	_ = process.currentTails[tailKey].tail.Stop()

	process.mu.Lock()
	delete(process.currentTails, tailKey)
	process.mu.Unlock()
}
