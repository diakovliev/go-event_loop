package event_loop

import (
    "log"
    "testing"
)

func TestCreateEventLoop(t *testing.T) {
    el := NewEventLoop("Test loop")
    if el == nil {
        t.Fatal("Can't create event loop")
    }

    h1_cntr := 0
    h2_cntr := 0

    el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler", func(e Event) bool {
        log.Printf("TEST1 %d", e.Type)
        h1_cntr += 1
        return true
    }))

    h2 := el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler 2", func(e Event) bool {
        log.Printf("TEST2 %d", e.Type)
        h2_cntr += 1
        return true
    }))

    if h2 == nil {
        t.Fatal("Unexpected Subscribe result: nil!")
    }

    el.Start()

    el.Send(NewEvent(EVENT_PRE_USER + 1, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 2, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 3, nil))

    if h2 != el.UnsubscribeSync(h2) {
        t.Fatal("Unexpected Unsubscribe result!")
    }

    el.Send(NewEvent(EVENT_PRE_USER + 4, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 5, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 6, nil))

    if el.StopAndJoin(0) != 0 {
        t.Fatal("Unexpected Stop result!")
    }

    if h1_cntr != 6 {
        t.Fatalf("Unexpected h1_cntr value: %d", h1_cntr)
    }
    if h2_cntr != 3 {
        t.Fatalf("Unexpected h1_cntr value: %d", h2_cntr)
    }
}

func TestCreateEventLoop2(t *testing.T) {
    el := NewEventLoop("Test loop")
    if el == nil {
        t.Fatal("Can't create event loop")
    }

    h1_cntr := 0
    h2_cntr := 0

    h1 := el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler", func(e Event) bool {
        log.Printf("TEST1 %d", e.Type)
        h1_cntr += 1
        return true
    }))

    if h1 == nil {
        t.Fatal("Unexpected Subscribe result: nil!")
    }

    el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler 2", func(e Event) bool {
        log.Printf("TEST2 %d", e.Type)
        h2_cntr += 1
        return true
    }))

    el.Start()

    el.Send(NewEvent(EVENT_PRE_USER + 1, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 2, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 3, nil))

    if h1 != el.UnsubscribeSync(h1) {
        t.Fatal("Unexpected Unsubscribe result!")
    }

    el.Send(NewEvent(EVENT_PRE_USER + 4, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 5, nil))
    el.Send(NewEvent(EVENT_PRE_USER + 6, nil))

    if el.StopAndJoin(0) != 0 {
        t.Fatal("Unexpected Stop result!")
    }

    if h1_cntr != 3 {
        t.Fatalf("Unexpected h1_cntr value: %d", h1_cntr)
    }
    if h2_cntr != 6 {
        t.Fatalf("Unexpected h1_cntr value: %d", h2_cntr)
    }
}

func TestSubscribe(t *testing.T) {
    el := NewEventLoop("Test loop")
    if el == nil {
        t.Fatal("Can't create event loop")
    }

    h1 := el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler", func(e Event) bool {
        log.Printf("TEST1 %d", e.Type)
        return false
    }))

    // Second subscribe of the same handler is valid
    el.Subscribe(h1)

    defer func() {
        r := recover()
        if r == nil {
            t.Fatal("No expected panic!")
        }
        log.Printf("Recovered from: %v", r)
    }()

    // Subscribing of different handler with the same name is invalid
    el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler", func(e Event) bool {
        log.Print("BAD HANDLER")
        return false
    }))
}

type TestExtSource struct {
    id string
    send bool
}
func (ts* TestExtSource) Id() string {
    return "test ext source " + ts.id
}
func (ts* TestExtSource) Get() []*Event {
    if (ts.send) {
        return nil
    }
    events := make([]*Event, 1)
    events[0] = NewEvent(EVENT_PRE_USER + 4, nil)
    ts.send = true
    return events
}
func (ts* TestExtSource) Empty() bool {
    return ts.send
}

func TestSource(t *testing.T) {
    el := NewEventLoop("Test loop")
    if el == nil {
        t.Fatal("Can't create event loop")
    }

    source1 := TestExtSource{id: "1", send: false}
    source2 := new(TestExtSource)
    source2.id = "2"
    source2.send = false

    el.AddSource(&source1)
    el.AddSource(source2)

    el.Subscribe(NewEventsHandler(EVENT_ALL, "test handler", func(e Event) bool {
        log.Printf("TEST1 %d", e.Type)
        return true
    }))

    el.Start()
    if el.StopAndJoin(1) != 1 {
        t.Fatal("Unexpected Stop result!")
    }

    if !source1.send {
        t.Fatal("Source1 is not used!")
    }
    if !source2.send {
        t.Fatal("Source2 is not used!")
    }
}


type TestExtSource2 struct {
    id string
    send bool
    cntr int
}
func (ts* TestExtSource2) Id() string {
    return "test ext source " + ts.id
}
func (ts* TestExtSource2) Get() []*Event {
    if ts.cntr < 3 {
        ts.cntr += 1
        return nil
    }

    if (ts.send) {
        return nil
    }
    events := make([]*Event, 1)
    events[0] = NewEvent(EVENT_PRE_USER + 4, nil)
    ts.send = true
    return events
}
func (ts* TestExtSource2) Empty() bool {
    return ts.send
}

func TestNoEventsHandler(t *testing.T) {
    loopHandlerCalled := false
    defaultHandlerCalled := false

    loopHandler := NewNoEventsHandler("test no events handler", func() {
        log.Printf("No events loop handler called")
        loopHandlerCalled = true
    })

    defaultHandler := NewDefaultEventsHandler("test not events default handler", func(e Event) bool {
        log.Printf("No events default handler called")
        defaultHandlerCalled = true
        return true
    })

    el := NewEventLoopExt("Test loop", defaultHandler, loopHandler)
    if el == nil {
        t.Fatal("Can't create event loop")
    }

    source1 := TestExtSource2{id: "X", send: false, cntr: 0}
    el.AddSource(&source1)

    el.Start()
    if el.StopAndJoin(1) != 1 {
        t.Fatal("Unexpected Stop result!")
    }

    if !defaultHandlerCalled {
        t.Fatal("Default handler not called!")
    }
    if !loopHandlerCalled {
        t.Fatal("Loop handler not called!")
    }
}
