package OperationServer

import (
	"github.com/Blizzardx/GoGameServer/Core/Event"
	"github.com/Blizzardx/GoGameServer/Core/Modules/Config"
	"github.com/Blizzardx/GoGameServer/Core/Modules/EventDefine"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/rpc"
	"github.com/davyxu/cellnet/timer"
	"github.com/davyxu/golog"
	"time"
)

type operationServer struct {
	log                      *golog.Logger
	gameServerMap            map[int32]*Network.ServerObject           //game server session 列表 key - server id ,value - server object
	singleComponentServerMap map[int32]*Network.ServerObject           //单例组件服务器 session 列表 key - server type ,value - server object
	multiComponentServerMap  map[int32]map[int32]*Network.ServerObject //多实例组件服务器 session 列表 key - server type ,value - server object list
	config                   *Config.NodeConfigInfo
}

func (server *operationServer) init() {
	server.log = golog.New("core.operationServer")
	server.gameServerMap = map[int32]*Network.ServerObject{}
	server.singleComponentServerMap = map[int32]*Network.ServerObject{}
	server.multiComponentServerMap = map[int32]map[int32]*Network.ServerObject{}
}

func (server *operationServer) Start(remoteConfigUrl string) {
	// 初始配置
	server.initConfig(remoteConfigUrl)
	//连接 其他game server
	server.connectGameServer()
	//连接单例组件服务器
	server.connectSingleComponentServer()
	//连接多实例组件服务器
	server.connectMultiComponentServer()

	//注册退出函数
	Network.RegisterExitCallback(server.onExitFunction)

	//事件通知服务器启动
	Event.KEventManager.Call(EventDefine.SystemEvent_ServerInit, nil)

	Network.Start()
}

//创建一个定时器
func (server *operationServer) CreateTicker(duration time.Duration, callback func(*timer.Loop)) {
	Network.CreateTicker(duration, callback)
}

func (server *operationServer) initConfig(remoteConfigUrl string) {

	// fetch config
	server.config = Config.FetchRemoteConfig(remoteConfigUrl)
}

// 入站(接收)的事件处理,切记，这里是在io线程中
func (server *operationServer) OnInboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
	var handled bool
	var err error

	inputEvent, handled, err = rpc.ResolveInboundEvent(inputEvent)

	if err != nil {
		server.log.Errorln("rpc.ResolveInboundEvent:", err)
		return
	}
	if handled {
		return inputEvent
	}
	msglog.WriteSendLogger(server.log, "ws", inputEvent.Session(), inputEvent.Message())
	return inputEvent
}

// 出站(发送)的事件处理，切记，这里是在io线程中
func (server *operationServer) OnOutboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
	handled, err := rpc.ResolveOutboundEvent(inputEvent)

	if err != nil {
		server.log.Errorln("rpc.ResolveOutboundEvent:", err)
		return nil
	}
	if handled {
		return inputEvent
	}
	msglog.WriteSendLogger(server.log, "ws", inputEvent.Session(), inputEvent.Message())
	return inputEvent
}

func (server *operationServer) onExitFunction() {

	//停止主动链接 gameserver
	for _, session := range server.gameServerMap {
		session.Stop()
	}
	server.gameServerMap = map[int32]*Network.ServerObject{}

	//停止主动链接 single component server
	for _, session := range server.singleComponentServerMap {
		session.Stop()
	}
	server.singleComponentServerMap = map[int32]*Network.ServerObject{}

	//停止主动链接 multi component server
	for _, serverMap := range server.multiComponentServerMap {
		for _, session := range serverMap {
			session.Stop()
		}
	}
	server.multiComponentServerMap = map[int32]map[int32]*Network.ServerObject{}

	//事件通知服务器退出
	Event.KEventManager.Call(EventDefine.SystemEvent_ServerQuit, nil)
}
