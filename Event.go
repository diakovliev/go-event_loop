package event_loop

import "fmt"

const (
    EVENT_ALL       = 0
    EVENT_PRE_USER  = 100
)

type IEventStringer interface {
    GetTypeString(uint64) string
}

type Event struct {
    Type uint64
    Payload interface{}
    stringer IEventStringer
}

func NewEventExt(t uint64, p interface{}, s IEventStringer) *Event {
    switch t {
    case EVENT_ALL, EVENT_PRE_USER:
        panic(fmt.Errorf("Can't make Event using special Event.Type value: %d!", t))
    }
    e := new(Event)
    e.Type      = t
    e.Payload   = p
    e.stringer  = s
    return e
}

func NewEvent(t uint64, p interface{}) *Event {
    e := NewEventExt(t, p, Event{})
    return e
}

func (e Event) GetTypeString(t uint64) string {
    switch t {
    case EVENT_ALL:
        return "ALL";
    case EVENT_PRE_USER:
        return "PRE";
    }
    if t > EVENT_PRE_USER {
        return "USER"
    } else {
        return "UNKNOWN"
    }
}

func (e Event) TypeString() string {
    if e.stringer != nil {
        return e.stringer.GetTypeString(e.Type)
    } else {
        return "UNKONWN"
    }
}
