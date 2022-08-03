package core

import (
	"sync"

	uuid "github.com/satori/go.uuid"
)

var EventDispatcher = &eventBusStruct{
	events: map[string]eventStruct{},
}

type EventData interface {
	Name() string
	Data() interface{}
}

type (
	eventCallbackInfo struct {
		id       string
		callback eventCallback
	}
	eventCallback  func(data interface{})
	eventStruct    map[int][]*eventCallbackInfo
	eventBusStruct struct {
		mu     sync.Mutex
		events map[string]eventStruct
	}
)

type EventDescription struct {
	Name     string
	Priority int
	Callback eventCallback
}

func (bus *eventBusStruct) Subscribe(description EventDescription) string {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, exists := bus.events[description.Name]; !exists {
		bus.events[description.Name] = eventStruct{}
	}
	if _, exists := bus.events[description.Name][description.Priority]; !exists {
		bus.events[description.Name][description.Priority] = []*eventCallbackInfo{}
	}
	eventUUID := uuid.NewV4().String()

	bus.events[description.Name][description.Priority] = append(bus.events[description.Name][description.Priority], &eventCallbackInfo{
		id:       eventUUID,
		callback: description.Callback,
	})

	return eventUUID
}

func (bus *eventBusStruct) Unsubscribe(subscribeID string) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	newEventList := map[string]eventStruct{}

	for k, events := range bus.events {
		newEventList[k] = eventStruct{}
		for k2, priorities := range events {
			newEventList[k][k2] = []*eventCallbackInfo{}
			for _, callbackInfo := range priorities {
				if callbackInfo.id != subscribeID {
					newEventList[k][k2] = append(newEventList[k][k2], callbackInfo)
				}
			}
		}
	}

	bus.events = newEventList
}

func (bus *eventBusStruct) Dispatch(event EventData) {
	if events, ok := bus.events[event.Name()]; ok {
		for _, priorities := range events {
			for _, callbackInfo := range priorities {
				go callbackInfo.callback(event.Data())
			}
		}
	}
}
