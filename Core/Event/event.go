package Event

import (
	"container/list"
	"errors"
	"github.com/davyxu/golog"
	"github.com/Blizzardx/GoGameServer/Core/Common"
)

var (
	KEventManager = &eventManager{}
	log           = golog.New("event manager")
)

func init() {
	KEventManager.events = map[string]*list.List{}
}

type KEvent struct {
	Name string
	Args interface{}
}

type KEventHandler struct {
	EventName   string
	EventWhen   func(ev *KEvent) bool
	EventAction func(ev *KEvent)
	EventNext   func()
}

func (self *KEventHandler) Name() string {
	return self.EventName
}
func (self *KEventHandler) When(ev *KEvent) bool {
	return self.EventWhen(ev)
}
func (self *KEventHandler) Action(ev *KEvent) {
	if self.EventWhen == nil || self.EventWhen(ev) {
		self.EventAction(ev)
		if self.EventNext != nil {
			self.EventNext()
		}
	}
}

func (self *KEventHandler) Next() {
	self.EventNext()
}

type eventManager struct {
	events map[string]*list.List
}

func (self *eventManager) AddSimpleKEvent(eventName string, action func(ev *KEvent)) (err error) {
	return self.AddKEvent(eventName, nil, action, nil)
}

func (self *eventManager) AddKEvent(eventName string, when func(ev *KEvent) bool, action func(ev *KEvent), next func()) (err error) {
	if eventName == "" || action == nil {
		return errors.New("arg error,action == nil,or eventName is nil")
	}

	event := &KEventHandler{
		EventName:   eventName,
		EventWhen:   when,
		EventAction: action,
		EventNext:   next,
	}

	some := KEventManager.events[event.Name()]
	if some == nil {
		some = list.New()
		KEventManager.events[event.Name()] = some
	}

	some.PushBack(event)
	return
}

func (self *eventManager) Call(name string, args interface{}) {
	some := KEventManager.events[name]

	if some == nil || name == "" {
		log.Errorln("event name is invalid")
		return
	}

	if some != nil {
		for e := some.Front(); e != nil; e = e.Next() {
			event := e.Value.(*KEventHandler)
			Common.SafeCall(func() {
				event.Action(&KEvent{Name: name, Args: args})
			})
		}
	}
}
