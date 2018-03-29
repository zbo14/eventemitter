package eventemitter

import (
	"fmt"
	"reflect"
	"testing"
)

func invalidParams(params ...interface{}) error {
	return fmt.Errorf("invalid params: %v", params)
}

func expect(t *testing.T, exp interface{}, res interface{}) {
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("expected %v got %v", exp, res)
	}
}

func TestEventEmitter(t *testing.T) {
	next := make(chan struct{})
	e := NewEventEmitter()
	sum := 0
	lis := func(e *EventEmitter, params ...interface{}) {
		temp := 0
		for _, param := range params {
			if x, ok := param.(int); ok {
				temp += x
			} else {
				e.Emit("error", invalidParams(params...))
				return
			}
		}
		sum += temp
		e.Emit("done")
	}
	e.On("add", lis)
	e.Once("done", func(_ *EventEmitter, _ ...interface{}) {
		expect(t, 1, len(e.ls["add"].on))
		expect(t, 0, len(e.ls["done"].once))
		expect(t, 6, sum)
		next <- struct{}{}
	})
	e.Emit("add", 1, 2, 3)
	<-next
	e.Once("error", func(_ *EventEmitter, params ...interface{}) {
		expect(t, 1, len(e.ls["error"].on))
		expect(t, 0, len(e.ls["error"].once))
		expect(t, 1, len(params))
		expect(t, invalidParams("foo", "bar"), params[0])
		expect(t, 6, sum)
		next <- struct{}{}
	})
	e.Emit("add", "foo", "bar")
	<-next
	e.Once("error", func(_ *EventEmitter, params ...interface{}) {
		expect(t, 1, len(e.ls["error"].on))
		expect(t, 0, len(e.ls["error"].once))
		expect(t, 1, len(params))
		expect(t, fmt.Errorf("unexpected event: sub"), params[0])
		expect(t, 6, sum)
		next <- struct{}{}
	})
	e.Emit("sub")
	<-next
	e.Once("error", func(_ *EventEmitter, params ...interface{}) {
		expect(t, 1, len(e.ls["error"].on))
		expect(t, 0, len(e.ls["error"].once))
		expect(t, 1, len(params))
		expect(t, fmt.Errorf("expected 'on' msg to have listener"), params[0])
		expect(t, 6, sum)
		next <- struct{}{}
	})
	e.On("blah", nil)
	<-next
	e.Once("error", func(_ *EventEmitter, params ...interface{}) {
		expect(t, 1, len(e.ls["error"].on))
		expect(t, 0, len(e.ls["error"].once))
		expect(t, 1, len(params))
		expect(t, fmt.Errorf("expected 'once' msg to have listener"), params[0])
		expect(t, 6, sum)
		e.Close()
		next <- struct{}{}
	})
	e.Once("bleh", nil)
	<-next
}
