package core

import (
	"reflect"
	"time"
)

const ProcessInfiniteLoopDelay = 100 * time.Millisecond

/*
 * Process interface.
 */
type ProcessInterface interface {
	Start()
	Stop()
	IsRunning() bool
	GetName() string
}

type RunningProcessInterface interface {
	Run() error
	HandleStop()
}

/*
 * Process struct.
 */
type ProcessStruct struct {
	RunningProcess RunningProcessInterface
	isRunning      bool
}

func (process *ProcessStruct) Start() {
	Logger.Infof(logPrefix, "Starting process %s", process.GetName())
	process.isRunning = true

	if err := process.RunningProcess.Run(); err != nil {
		Logger.Errorf(logPrefix, "Fatal error for process %s : %s", process.GetName(), err)
	}

	Logger.Infof(logPrefix, "Shutting down process %s", process.GetName())
	process.isRunning = false
}

func (process *ProcessStruct) Stop() {
	if process.isRunning {
		process.RunningProcess.HandleStop()
	}
}

func (process *ProcessStruct) IsRunning() bool {
	return process.isRunning
}

func (process *ProcessStruct) GetName() string {
	return reflect.Indirect(reflect.ValueOf(process.RunningProcess)).Type().String()
}

func ProcessInfiniteLoop(delay time.Duration, exitChan chan bool, handler func()) {
	askForStop := false
	var lastExecution time.Time

	go func() {
		askForStop = <-exitChan
	}()

	for !askForStop {
		if time.Now().Before(lastExecution.Add(delay)) {
			time.Sleep(ProcessInfiniteLoopDelay)

			continue
		}
		lastExecution = time.Now()

		// run
		handler()
	}
}
