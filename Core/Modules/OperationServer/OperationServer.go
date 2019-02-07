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
	log           *golog.Logger
	gameServerMap map[int32]*Network.ServerObject //game server session 列表 key - server id ,value - server object
	config        *Config.NodeConfigInfo
}

func (server *operationServer) init() {
	server.log = golog.New("core.operationServer")
}

func (server *operationServer) Start() {
	// 初始配置
	server.initConfig()
	//连接 其他game server
	server.connectGameServer()

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

func (server *operationServer) initConfig() {

	// fetch config
	server.config = Config.FetchRemoteConfig("")
}

func (server *operationServer) connectGameServer() {
	for _, serverInfo := range server.config.GameServerList {
		serverObj := Network.ConnectToTCP(serverInfo.InternalAddress+":"+serverInfo.GameServerListenPort, server.onConnectServerSessionChange, server, serverInfo)
		server.gameServerMap[serverInfo.LogicId] = serverObj
	}
}

//主动连接： gameServer 连接发生变化
func (server *operationServer) onConnectServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	serverConfig := arg.(*Config.NodeConfigGameServerInfo)
	if eventName == "SessionConnected" {
		server.onGameServerServerConnected(serverConfig.LogicId, ev.Session())

	} else if eventName == "SessionClosed" {
		server.onGameServerServerDisConnect(serverConfig.LogicId)
	}
}

func (server *operationServer) onGameServerServerConnected(serverType int32, session cellnet.Session) {
	if serverObj, ok := server.gameServerMap[serverType]; ok {
		serverObj.Session = session
	}
}
func (server *operationServer) onGameServerServerDisConnect(serverType int32) {
	if serverObj, ok := server.gameServerMap[serverType]; ok {
		serverObj.Session = nil
	}
}

//给GameServer 发送消息
func (server *operationServer) sendMessageToGameServer(gsId int32, msg interface{}) {
	if session, ok := server.gameServerMap[gsId]; ok {
		if session.Session == nil {
			return
		}
		//
		session.Session.Send(msg)
	}
}

//给GameServer 发送异步RPC 消息
func (server *operationServer) sendRPCMessageToGameServer(gsId int32, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	if session, ok := server.gameServerMap[gsId]; ok {
		if session.Session == nil {
			return
		}
		//
		Network.RPC(session.Session, msg, timeOut, userCallback)
	}
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

	//事件通知服务器退出
	Event.KEventManager.Call(EventDefine.SystemEvent_ServerQuit, nil)
}
