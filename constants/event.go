package constants

import (
	"sync"
)

type eventItem struct {
	params []interface{}
}
type eventCB struct {
	cb_id int
	cb    func(args ...interface{})
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////
type IEventManager interface {
	AddEventLister(cb func(args ...interface{})) *eventCB
	RemoveEventLister(c *eventCB)

	DispatchEvent(params ...interface{})
}

type eventManager struct {
	cb_id         int
	events_cb_map sync.Map //*eventCB
	event_chan    chan *eventItem
}

func (m *eventManager) AddEventLister(cb func(args ...interface{})) *eventCB {
	m.cb_id++
	evtcb := &eventCB{
		cb_id: m.cb_id,
		cb:    cb,
	}
	m.events_cb_map.Store(evtcb.cb_id, evtcb)

	return evtcb
}
func (m *eventManager) RemoveEventLister(c *eventCB) {
	if c == nil {
		return
	}

	m.events_cb_map.Delete(c.cb_id)
}
func (m *eventManager) DispatchEvent(params ...interface{}) {
	t := &eventItem{
		params: append([]interface{}{}, params...),
	}

	m.event_chan <- t
}

func (m *eventManager) run() {
	for evt := range m.event_chan {
		if evt != nil {
			m.events_cb_map.Range(func(key, value any) bool {
				cb, ok := value.(*eventCB)
				if ok && cb != nil {
					go func(cb *eventCB, evt *eventItem) {
						defer Recover()()
						cb.cb(evt.params...)
					}(cb, evt)
				}
				return true
			})
		}
	}
}
func NewEventManager() IEventManager {
	t := &eventManager{
		cb_id:         0,
		events_cb_map: sync.Map{},
		event_chan:    make(chan *eventItem, 1024),
	}
	go func() {
		defer Recover()()
		t.run()
	}()
	return t
}
