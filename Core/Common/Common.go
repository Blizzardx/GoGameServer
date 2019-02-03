package Common

import (
	"github.com/davyxu/golog"
	"runtime/debug"
)

var (
	log = golog.New("core.common")
)

/**
安全的运行func
*/
func SafeCall(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln(string(debug.Stack()))
		}
	}()
	f()
}
