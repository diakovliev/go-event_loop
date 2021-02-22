package event_loop

import "sync"

type Handlers struct {
    values []*EventsHandler
    guard sync.Mutex
}

func NewHandlers() *Handlers {
    h := new(Handlers)
    h.values = make([]*EventsHandler, 0, 100)
    return h
}

func (h *Handlers) Store(name string, handler *EventsHandler) {
    h.guard.Lock()
    defer h.guard.Unlock()

    h.values = append(h.values, handler)
}

func (h *Handlers) Load(name string) (*EventsHandler, bool) {
    h.guard.Lock()
    defer h.guard.Unlock()

    for _, handler := range(h.values) {
        if handler.Name == name {
            return handler, true
        }
    }

    return nil, false
}

func (h *Handlers) Delete(name string) {
    h.guard.Lock()
    defer h.guard.Unlock()

    idx := -1

    for iter_idx, handler := range(h.values) {
        if handler.Name == name {
            idx = iter_idx
            break
        }
    }

    if idx != -1 {
        h.values = append(h.values[:idx], h.values[idx+1:]...)
    }
}

func (h *Handlers) Range(callback func (key, value interface{}) bool) {
    h.guard.Lock()
    defer h.guard.Unlock()

    for _, handler := range(h.values) {
        if !callback(handler.Name, handler) {
            break
        }
    }
}
