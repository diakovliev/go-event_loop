package event_loop

import (
    "sync"
    "log"
)

type ISource interface {

    Id() string

    // Method must return nil if no events.
    Get() []*Event

    Empty() bool
}

type Sources struct {
    id string

    sources []ISource
    guard sync.Mutex
}

func NewSources(id string) *Sources {
    ss := new(Sources)
    ss.id       = id
    ss.sources  = make([]ISource, 0, 100)
    return ss
}

func (ss *Sources) Add(source ISource) {
    ss.guard.Lock()
    defer ss.guard.Unlock()

    ss.sources = append(ss.sources, source)
}

func (ss *Sources) Id() string {
    return ss.id
}

func (ss *Sources) Get() []*Event {
    ss.guard.Lock()
    defer ss.guard.Unlock()

    events := make([]*Event, 0, 100)
    for _, source := range(ss.sources) {
        source_events := source.Get()
        if source_events != nil {
            log.Printf("Reveive events from source '%s'.", source.Id())
            events = append(events, source_events...)
        }
    }

    if len(events) == 0 {
        return nil
    }

    return events
}

func (ss *Sources) Empty() bool {
    ss.guard.Lock()
    defer ss.guard.Unlock()

    for _, source := range(ss.sources) {
        if !source.Empty() {
            log.Printf("Source '%s' is not empty.", source.Id())
            return false
        }
    }

    return true
}
