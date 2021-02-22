package event_loop

import (
    "fmt"
    "errors"
    "log"
    "time"
    "sync/atomic"
)

const (
    internal_EVENT_LOOP_STARTED = 0
    internal_EVENT_LOOP_STOPPED = 1
)

type eventLoopChannel struct {
    C chan int
}

func newEventLoopChannel() *eventLoopChannel {
    c := new(eventLoopChannel)
    c.C = make(chan int)
    return c
}


type EventLoop struct {
    Name string

    stopValue atomic.Value
    channel *eventLoopChannel

    queue *Queue

    sources *Sources
    handlers *Handlers
    defaultHandler *EventsHandler
    loopHandler *NoEventsHandler
}

func NewEventLoopExt(name string, defaultHandler *EventsHandler, loopHandler *NoEventsHandler) *EventLoop {
    el := new(EventLoop)

    el.channel          = nil

    el.Name             = name
    el.queue            = newQueue()
    el.sources          = NewSources(name + ".sources")
    el.handlers         = NewHandlers()
    el.defaultHandler   = defaultHandler
    el.loopHandler      = loopHandler

    log.Printf("Created event loop: %s", el.Name)
    return el
}

func NewEventLoop(name string) *EventLoop {
    return NewEventLoopExt(name, nil, nil)
}


func (el *EventLoop) Subscribe(handler *EventsHandler) *EventsHandler {
    h, ok := el.handlers.Load(handler.Name)
    if ok {
        if h != handler {
            panic(fmt.Errorf("Different handler with name '%s' already subscribed for event loop '%s'!", handler.Name, el.Name))
        } else {
            log.Printf("%s: Handler '%s' already subscribed.", el.Name, handler.Name)
            return handler
        }
    }

    el.handlers.Store(handler.Name, handler)
    log.Printf("%s: Handler '%s' subscribed.", el.Name, handler.Name)

    return handler
}

func (el *EventLoop) Unsubscribe(handler *EventsHandler) *EventsHandler {
    _, ok := el.handlers.Load(handler.Name)
    if !ok {
        log.Printf("%s: Handler '%s' is not subsribed.", el.Name, handler.Name)
        return handler
    }

    el.handlers.Delete(handler.Name)
    log.Printf("%s: Handler '%s' unsubscribed.", el.Name, handler.Name)

    return handler
}

func (el *EventLoop) SetDefaultHandler(handler *EventsHandler) *EventLoop {
    if el.defaultHandler != nil && el.defaultHandler != handler {
        panic(fmt.Errorf("Different default handler '%s' already set for the event loop '%s'!", el.defaultHandler.Name, el.Name))
        return nil
    }

    log.Printf("%s: Default handler '%s' is set.", el.Name, handler.Name)

    el.defaultHandler = handler

    return el
}

func (el *EventLoop) SetLoopHandler(handler *NoEventsHandler) *EventLoop {
    if el.loopHandler != nil && el.loopHandler != handler {
        panic(fmt.Errorf("Different loop handler '%s' already set for the event loop '%s'!", el.loopHandler.Name, el.Name))
        return nil
    }

    log.Printf("%s: Loop handler '%s' is set.", el.Name, handler.Name)

    el.loopHandler = handler

    return el
}

func (el *EventLoop) AddSource(source ISource) {
    log.Printf("%s: Add event source '%s'", el.Name, source.Id())
    el.sources.Add(source)
}


func (el *EventLoop) handle(event Event) int {
    handled := 0

    el.handlers.Range(func(key, value interface{}) bool {
        handler, ok := value.(*EventsHandler)
        if !ok {
            panic(fmt.Errorf("Not an *EventsHandler in handlers map!"))
        }

        ret := true

        if handler.Type == EVENT_ALL || event.Type == handler.Type {
            log.Printf("%s: handle %s(%d) by handler '%s'", el.Name, event.TypeString(), event.Type, handler.Name)
            ret = handler.callCallback(event)
            log.Printf("%s: callback result: %v", el.Name, ret)
            handled += 1
        }

        return ret
    })

    if handled == 0 && el.defaultHandler != nil {
        log.Printf("%s: handle %s(%d) by default handler '%s'", el.Name, event.TypeString(), event.Type, el.defaultHandler.Name)
        el.defaultHandler.callCallback(event)
    }

    return handled
}

func (el *EventLoop) loopHandle() {
    if el.loopHandler == nil {
        //log.Printf("%s: No loop handler.", el.Name)
        time.Sleep(10 * time.Millisecond)
        return
    }

    log.Printf("%s: Call loop handler '%s'.", el.Name, el.loopHandler.Name)
    el.loopHandler.call()
}

func (el *EventLoop) getEvent() *Event {
    ext_events := el.sources.Get()
    if ext_events == nil {
        return el.queue.Dequeue()
    }

    for _, event := range(ext_events) {
        el.queue.Enqueue(event)
    }

    return el.queue.Dequeue()
}

func (el *EventLoop) Empty() bool {
    return el.queue.Empty() && el.sources.Empty()
}

func (el *EventLoop) run() *EventLoop {
    log.Printf("%s: Enter event loop", el.Name)
    if el.channel != nil {
        el.channel.C <- internal_EVENT_LOOP_STARTED
    }

    defer func() {
        log.Printf("%s: Leave event loop", el.Name)
        if el.channel != nil {
            el.channel.C <- internal_EVENT_LOOP_STOPPED
        }
    }()

    var stop bool = false

    for {

        is_stop_set := el.stopValue.Load()
        if is_stop_set != nil {
            stop = true
        }

        if stop {
            if el.channel == nil || el.Empty() {
                log.Printf("%s: Leave event loop...", el.Name)
                break
            } else {
                log.Printf("%s: Waiting for pending events processing...", el.Name)
            }
        }

        event := el.getEvent()
        if event != nil {
            handled := el.handle(*event)
            if handled == 0 {
                log.Printf("%s: No events handlers for event type %s(%d), ignore event.", el.Name, event.TypeString(), event.Type)
            }
        } else {
            el.loopHandle()
        }
    }

    log.Printf("%s: Stopped loop.", el.Name)

    return el;
}

func (el *EventLoop) Run() {
    el.run()
}

// Sync method will try to wait the moment when EventLoop queue will be empty. Waiting is simple
// loop with sleepeng 100ms in them. The maxIterations allows to control max waiting time. The
// maxIterations == 0 means not limited waiting. Method will return result of queue.Empty() call
// after waiting.
func (el *EventLoop) Sync(maxIterations int) bool {
    if el.channel == nil {
        panic(fmt.Errorf("Sync called for non joinable event loop '%s'!", el.Name))
        return false
    }

    log.Printf("Sync with event loop '%s'. Max iterations: %d...", el.Name, maxIterations)

    cntr := 0
    for !el.Empty() {
        time.Sleep(100 * time.Millisecond)
        cntr += 1
        if maxIterations > 0 && cntr >= maxIterations {
            break
        }
    }

    ret := el.Empty()
    log.Printf("Sync with event loop '%s'. Result: %v.", el.Name, ret)

    return ret
}

// Mainly for testing, avoid using this in real code.
func (el *EventLoop) UnsubscribeSync(handler *EventsHandler) *EventsHandler {
    el.Sync(0)
    return el.Unsubscribe(handler)
}


func (el *EventLoop) Start() *EventLoop {
    log.Printf("%s: Start event loop.", el.Name)
    el.channel = newEventLoopChannel()
    go el.run()
    ret := <- el.channel.C
    if ret != internal_EVENT_LOOP_STARTED {
        panic(fmt.Errorf("Got wrong value from channel! Value: %d", ret))
    }
    return el
}

func (el *EventLoop) Join() int {
    if el.channel == nil {
        panic(fmt.Errorf("Join called for non joinable event loop '%s'!", el.Name))
        return 0
    }

    cret := <- el.channel.C
    if cret != internal_EVENT_LOOP_STOPPED {
        panic(fmt.Errorf("Got wrong value from channel! Value: %d", cret))
        return 0
    }

    iret := el.stopValue.Load()
    if iret == nil {
        panic(errors.New("No stop valus is set!"))
        return 0
    }

    ret, ok := iret.(int)
    if !ok {
        panic(errors.New("Not an integer!"))
        return 0
    }

    return ret
}

func (el *EventLoop) Stop(ret int) {
    log.Printf("%s: Stop event loop.", el.Name)
    el.stopValue.Store(ret)
}

func (el *EventLoop) StopAndJoin(ret int) int {
    el.Stop(ret)
    return el.Join()
}

func (el *EventLoop) Send(event *Event) *EventLoop {
    log.Printf("%s: Send event %s(%d) to event loop.", el.Name, event.TypeString(), event.Type)
    el.queue.Enqueue(event)
    return el
}
