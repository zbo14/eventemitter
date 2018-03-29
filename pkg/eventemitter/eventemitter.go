package eventemitter

import "fmt"

type listener func(*EventEmitter, ...interface{})

type listeners struct {
	on   []listener
	once []listener
}

type message struct {
	cmd    string
	event  string
	l      listener
	params []interface{}
}

func (msg *message) String() string {
	return fmt.Sprintf("message{cmd=%s,event=%s,params=%v}", msg.cmd, msg.event, msg.params)
}

// EventEmitter receives messages from the channel, sets event listeners, and fires events
type EventEmitter struct {
	ch chan *message
	ls map[string]*listeners
}

func defaultErrListener(_ *EventEmitter, params ...interface{}) {
	if len(params) == 1 {
		if err, ok := params[0].(error); ok {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
}

// NewEventEmitter returns an EventEmitter
func NewEventEmitter() *EventEmitter {
	e := &EventEmitter{
		ch: make(chan *message),
		ls: make(map[string]*listeners),
	}
	e.on("error", defaultErrListener)
	go e.listen()
	return e
}

// Close closes the EventEmitter channel
func (e *EventEmitter) Close() {
	close(e.ch)
}

func (e *EventEmitter) listen() {
	for msg, more := <-e.ch; more; msg, more = <-e.ch {
		switch msg.cmd {
		case "on":
			e.handleOn(msg)
		case "once":
			e.handleOnce(msg)
		case "emit":
			e.handleEmit(msg)
		default:
			e.Emit("error", fmt.Errorf("unexpected cmd: %s", msg.cmd))
		}
	}
}

func (e *EventEmitter) on(ev string, l listener) {
	if e.ls[ev] == nil {
		e.ls[ev] = new(listeners)
	}
	e.ls[ev].on = append(e.ls[ev].on, l)
}

func (e *EventEmitter) once(ev string, l listener) {
	if e.ls[ev] == nil {
		e.ls[ev] = new(listeners)
	}
	e.ls[ev].once = append(e.ls[ev].once, l)
}

func (e *EventEmitter) fire(ev string, params ...interface{}) {
	ls := e.ls[ev]
	if len(ls.on) > 0 {
		for _, l := range ls.on {
			go l(e, params...)
		}
	}
	if len(ls.once) > 0 {
		for _, l := range ls.once {
			go l(e, params...)
		}
		ls.once = ls.once[:0]
	}
}

func (e *EventEmitter) handleOn(msg *message) {
	if msg.l == nil {
		e.fire("error", fmt.Errorf("expected 'on' msg to have listener"))
	} else {
		e.on(msg.event, msg.l)
	}
}

func (e *EventEmitter) handleOnce(msg *message) {
	if msg.l == nil {
		e.fire("error", fmt.Errorf("expected 'once' msg to have listener"))
	} else {
		e.once(msg.event, msg.l)
	}
}

func (e *EventEmitter) handleEmit(msg *message) {
	if e.ls[msg.event] == nil {
		e.fire("error", fmt.Errorf("unexpected event: %s", msg.event))
	} else {
		e.fire(msg.event, msg.params...)
	}
}

// On sends an "on" message with event and listener on the EventEmitter channel
func (e *EventEmitter) On(ev string, l listener) {
	e.ch <- &message{
		cmd:   "on",
		event: ev,
		l:     l,
	}
}

// Once sends a "once" message with event and listener on the EventEmitter channel
func (e *EventEmitter) Once(ev string, l listener) {
	e.ch <- &message{
		cmd:   "once",
		event: ev,
		l:     l,
	}
}

// Emit sends an "emit" message with event and params on the EventEmitter channel
func (e *EventEmitter) Emit(ev string, params ...interface{}) {
	e.ch <- &message{
		cmd:    "emit",
		event:  ev,
		params: params,
	}
}
