package event_loop

import (
    "log"
)

type EventsHandler struct {
    Name string
    Type uint64
    callback func(Event) bool
    async bool
}

func NewEventsHandlerExt(t uint64, name string, callback func(Event) bool, async bool) *EventsHandler {
    h := new(EventsHandler)
    h.Type      = t
    h.Name      = name
    h.callback  = callback
    h.async     = async
    return h
}

func NewEventsHandler(t uint64, name string, callback func(Event) bool) *EventsHandler {
    return NewEventsHandlerExt(t, name, callback, false)
}

func NewDefaultEventsHandler(name string, callback func(Event) bool) *EventsHandler {
    return NewEventsHandlerExt(EVENT_ALL, name, callback, false)
}

func (h *EventsHandler) callCallback(event Event) bool {
    if h.async {
        log.Printf("%s: Call callback asynchoniously.", h.Name)
        go h.callback(event)
        return true
    }

    log.Printf("%s: Call callback.", h.Name)
    result := h.callback(event)
    if !result {
        log.Printf("%s: Event %s(%d) processing complete.", h.Name, event.TypeString(), event.Type)
    }
    return result
}


type NoEventsHandler struct {
    Name string
    callback func()
}

func NewNoEventsHandler(name string, callback func()) *NoEventsHandler {
    h := new(NoEventsHandler)
    h.Name      = name
    h.callback  = callback
    return h
}

func (h *NoEventsHandler) call() {
    h.callback()
}
