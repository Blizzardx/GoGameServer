package GameServer

import (
	"github.com/Blizzardx/GoGameServer/Core/Common"
	"github.com/Blizzardx/GoGameServer/Core/Event"
	"github.com/Blizzardx/GoGameServer/Core/InternalMessage"
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

type OnPlayerDisconnect func(playerId string)
type OnKickPlayer func(playerId string)

type GameServer struct {
	log                       *golog.Logger
	playerIdToSession         map[string]cellnet.Session
	sessionIdToPlayerId       map[int64]string
	gameServerMap             map[int32]*Network.ServerObject           //game server session 列表 key - server id ,value - server object
	singleComponentServerMap  map[int32]*Network.ServerObject           //单例组件服务器 session 列表 key - server type ,value - server object
	multiComponentServerMap   map[int32]map[int32]*Network.ServerObject //多实例组件服务器 session 列表 key - server type ,value - server object list
	transmitMessageMap        map[string]int32
	config                    *Config.NodeConfigInfo
	selfConfig                *Config.NodeConfigGameServerInfo
	logicId                   int32
	onKickPlayerHandler       OnKickPlayer
	onPlayerDisconnectHandler OnPlayerDisconnect
	listenClientSession       *Network.ServerObject
	listenGameServerSession   *Network.ServerObject
	messageInterceptorHandler map[string]Network.TransmitMessageInterceptor
}
type PlayerManager interface {
	OnPlayerDisconnect(playerId string)
	KickPlayer(playerId string)
}

func CreateGameServer() *GameServer {
	server := &GameServer{}
	server.init()
	return server
}
func (server *GameServer) RegisterPlayerConnectHandler(kickHandler OnKickPlayer, onPlayerDisconnectHandler OnPlayerDisconnect) {
	server.onKickPlayerHandler = kickHandler
	server.onPlayerDisconnectHandler = onPlayerDisconnectHandler
}
func (server *GameServer) init() {
	server.log = golog.New("core.GameServer")
	server.playerIdToSession = map[string]cellnet.Session{}
	server.sessionIdToPlayerId = map[int64]string{}
	server.gameServerMap = map[int32]*Network.ServerObject{}
	server.singleComponentServerMap = map[int32]*Network.ServerObject{}
	server.multiComponentServerMap = map[int32]map[int32]*Network.ServerObject{}
	server.transmitMessageMap = map[string]int32{}
	server.messageInterceptorHandler = map[string]Network.TransmitMessageInterceptor{}
}
func (server *GameServer) Start(remoteConfigUrl string, serverId int32) {
	server.logicId = serverId
	// 初始配置
	server.initConfig(remoteConfigUrl)
	//监听game server 连接
	server.listenGameServer()
	//连接 其他game server
	server.connectGameServer()
	//连接单例组件服务器
	server.connectSingleComponentServer()
	//连接多实例组件服务器
	server.connectMultiComponentServer()
	//监听 客户端 连接
	if !server.listenClient() {
		return
	}

	//注册退出函数
	Network.RegisterExitCallback(server.onExitFunction)

	//事件通知服务器启动
	Event.KEventManager.Call(EventDefine.SystemEvent_ServerInit, nil)

	Network.Start()
}
func (server *GameServer) Post(f func()) {
	Network.Post(f)
}

//注册消息处理器
func (server *GameServer) RegisterMessage(messageName string, callback cellnet.EventCallback) {
	Network.RegisterMessage(messageName, callback)
}

//注册消息转发 只能注册 客户端消息 到 哪个单例组件服务器
func (server *GameServer) RegisterMessageTransmit(messageName string, toServerType int32) {
	if _, ok := server.transmitMessageMap[messageName]; ok {
		server.log.Errorf("already register message transmit ", messageName, toServerType)
		return
	}
	server.transmitMessageMap[messageName] = toServerType
}

//注册消息拦截器 在io层拦截 并停止消息传递
func (server *GameServer) RegisterMessageInterceptor(messageName string, handler Network.TransmitMessageInterceptor) {
	if _, ok := server.messageInterceptorHandler[messageName]; ok {
		server.log.Errorf("already register message interceptor ", messageName, handler)
		return
	}
	server.messageInterceptorHandler[messageName] = handler
}

//创建一个定时器
func (server *GameServer) CreateTicker(duration time.Duration, callback func(*timer.Loop)) {
	Network.CreateTicker(duration, callback)
}
func (server *GameServer) GetNodeConfig() *Config.NodeConfigInfo {
	return server.config
}
func (server *GameServer) GetServerId() int32 {
	return server.logicId
}
func (server *GameServer) beginConnectOtherServer() {
	var serverList []string
	for _, serverAdd := range serverList {
		Network.ConnectToTCP(serverAdd, server.onConnectServerSessionChange, server, nil)
	}
}

// 入站(接收)的事件处理,切记，这里是在io线程中
func (server *GameServer) OnInboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
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

	switch msg := inputEvent.Message().(type) {
	case *InternalMessage.InternalMsg:
		//这类包装过的内部消息 直接发给客户端
		server.sendInternalMessageToClient(msg)
		msglog.WriteRecvLogger(server.log, "ws", inputEvent.Session(), inputEvent.Message())
		return nil
	default:
		//先检查是不是需要转发的消息
		meta := cellnet.MessageMetaByMsg(inputEvent.Message())
		if nil == meta {
			//
			server.log.Errorf("not found message meta by name", meta.FullName())
			return nil
		}
		msgName := meta.FullName()
		if transmitMsgHandler, ok := server.messageInterceptorHandler[msgName]; ok {
			Common.SafeCall(func() {
				transmitMsgHandler(inputEvent.Session().ID(), msgName, inputEvent.Message())
			})
			return nil
		}

		if targetServerType, ok := server.transmitMessageMap[msgName]; ok {
			//检查这根连接是不是有客户端id
			playerId := server.getPlayerIdBySession(inputEvent.Session().ID())
			if playerId == "" {
				server.log.Errorf("error transmit message ,player not login ", msgName)
				return nil
			}
			//直接转发消息
			data, err := meta.Codec.Encode(inputEvent.Message(), nil)
			if err != nil {
				server.log.Errorf("error on decode msg by name", msgName)
				return nil
			}
			internalMsg := &InternalMessage.InternalMsg{MessageId: meta.ID, MessageBody: data.([]byte), PlayerId: playerId}
			server.SendMessageToSingleComponentServer(targetServerType, internalMsg)
			msglog.WriteRecvLogger(server.log, "ws", inputEvent.Session(), internalMsg)
			return nil
		}
		msglog.WriteRecvLogger(server.log, "ws", inputEvent.Session(), inputEvent.Message())
		return inputEvent
	}

	msglog.WriteRecvLogger(server.log, "ws", inputEvent.Session(), inputEvent.Message())
	return inputEvent
}

// 出站(发送)的事件处理，切记，这里是在io线程中
func (server *GameServer) OnOutboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
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

func (server *GameServer) onExitFunction() {

	//停止客户端监听
	if nil != server.listenClientSession {
		server.listenClientSession.Stop()
		server.listenClientSession = nil
	}

	//停止game server监听
	if nil != server.listenGameServerSession {
		server.listenGameServerSession.Stop()
		server.listenGameServerSession = nil
	}

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
