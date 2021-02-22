package event_loop

import (
    "sync"
)

type Queue struct {
    events []*Event
    guard sync.Mutex
}

func newQueue() *Queue {
    result := new(Queue)
    result.events = make([]*Event, 0, 100)
    return result
}

func (q *Queue) Empty() bool {
    return len(q.events) == 0
}

func (q *Queue) Enqueue(event *Event) {
    q.guard.Lock()
    defer q.guard.Unlock()

    q.events = append(q.events, event)
}

func (q *Queue) Dequeue() *Event {
    q.guard.Lock()
    defer q.guard.Unlock()

    if q.Empty() {
        return nil
    }

    event := q.events[0]
    q.events = q.events[1:]
    return event
}
