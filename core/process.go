package core

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
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

func RunProcesses(processes []ProcessInterface) {
	nbCtrlC := 0
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	// Catch ctrl+c signal
	go func(processes []ProcessInterface) {
		for range sigChan {
			nbCtrlC++

			if nbCtrlC >= 2 {
				Logger.Infof(logPrefix, "Force exit !")
				os.Exit(0)
			}
			Logger.Infof(logPrefix, "Receive stop signal : kill process properly ... press Ctrl+c again to force kill")

			// Shutdown all running processes
			for _, process := range processes {
				process.Stop()
			}
		}
	}(processes)

	// Start services
	c := make(chan bool)
	for _, process := range processes {
		go func(process ProcessInterface) {
			defer func() {
				if r := recover(); r != nil {
					Logger.Criticalf(process.GetName(), "Panic occurred : %v", r)
				}
			}()
			process.Start()
			c <- true
		}(process)
	}

	// Wait for all processes exited
	for i := 0; i < len(processes); i++ {
		<-c
	}

	Logger.Infof(logPrefix, "All processes exited")
}
