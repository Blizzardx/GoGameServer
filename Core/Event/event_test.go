package Event

import "testing"

type A struct {
	Name string
}

func Test_event(t *testing.T) {

	KEventManager.AddKEvent("say", nil, func(event *KEvent) {
		a := event.Args.([]string)
		log.Debugln(a)
	}, func() {

	})

	KEventManager.Call("say", []string{"123", "222"})
}
