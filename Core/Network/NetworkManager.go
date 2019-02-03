package Network

import (
	"github.com/davyxu/cellnet"
	_ "github.com/davyxu/cellnet/codec/json"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/gorillaws"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/gorillaws"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/cellnet/timer"
	"github.com/davyxu/golog"
	"github.com/Blizzardx/GoGameServer/Core/Common"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

type OnSessionChange func(eventName string, ev cellnet.Event, arg interface{})

type TransmitMessageInterceptor func(sessionId int64, messageName string, messageBody interface{})

var (
	log                  *golog.Logger
	MainLogicQueue       = cellnet.NewEventQueue()
	handlers             = map[string]cellnet.EventCallback{}
	tickerLoop           []*timer.Loop
	onExitServerCallback func()
)

func init() {
	log = golog.New("core.server")
}

func Start() {
	//捕获主线程的异常
	MainLogicQueue.EnableCapturePanic(true)
	// 事件队列开始循环
	MainLogicQueue.StartLoop()

	// 阻塞等待事件队列结束退出( 在另外的goroutine调用queue.StopLoop() )
	go Common.SafeCall(exitFunc)

	log.Infof("server is started!")
	MainLogicQueue.Wait()
}

func Post(f func()) {
	if f == nil {
		log.Errorln(string(debug.Stack()))
		return
	}
	MainLogicQueue.Post(f)
}

//注册消息处理器
func RegisterMessage(messageName string, callback cellnet.EventCallback) {
	handlers[messageName] = callback
}

//创建一个定时器
func CreateTicker(duration time.Duration, callback func(*timer.Loop)) {
	ticker := timer.NewLoop(MainLogicQueue, duration, callback, nil)
	tickerLoop = append(tickerLoop, ticker)
	ticker.Start()
}

func RegisterExitCallback(callback func()) {
	onExitServerCallback = callback
}

func exitFunc() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Infof("exit %s begin")
	if nil != onExitServerCallback {
		Common.SafeCall(onExitServerCallback)
	}
	if tickerLoop != nil {
		for _, loop := range tickerLoop {
			loop.Stop()
		}
	}
	MainLogicQueue.StopLoop()
	log.Infof("exit %s over")

}
func createPipeline(processorName string, peerType string, address string, sessionChangeAction OnSessionChange, hooker cellnet.EventHooker, arg interface{}) *ServerObject {
	serverObject := &ServerObject{}
	serverObject.peer = peer.NewGenericPeer(peerType, "client", address, MainLogicQueue)
	serverObject.dispatcher = proc.NewMessageDispatcherBindPeer(serverObject.peer, processorName)
	serverObject.dispatcher.RegisterMessage("cellnet.SessionConnected", func(ev cellnet.Event) {
		//serverObject.Session = ev.Session()
		log.Infoln("与", address, "建立连接")
		sessionChangeAction("SessionConnected", ev, arg)
	})
	serverObject.dispatcher.RegisterMessage("cellnet.SessionClosed", func(ev cellnet.Event) {
		// 会话断开时
		log.Errorln("与", address, "断开连接了！！！")
		sessionChangeAction("SessionClosed", ev, arg)
	})

	for k, v := range handlers {
		serverObject.dispatcher.RegisterMessage(k, v)
	}

	serverObject.peer.(proc.ProcessorBundle).SetHooker(hooker)
	serverObject.peer.Start()

	return serverObject
}
