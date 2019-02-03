package ComponentServer

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/rpc"
	"github.com/davyxu/cellnet/timer"
	"github.com/davyxu/golog"
	"github.com/Blizzardx/GoGameServer/Core/Common"
	"github.com/Blizzardx/GoGameServer/Core/InternalMessage"
	"github.com/Blizzardx/GoGameServer/Core/Modules/Config"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"reflect"
	"time"
)

type CustomCellnetEvent struct {
	cellnet.RecvMsgEvent
	PlayerId string
}
type ComponentServer struct {
	log                       *golog.Logger
	gameServerIdToSession     map[int32]cellnet.Session
	sessionIdToGameServerId   map[int64]int32
	config                    *Config.NodeConfigInfo
	selfConfig                *Config.NodeConfigComponentServerInfo
	logicId                   int32
	instanceId                int32
	isSingleInstance          bool
	listenGameServerSession   *Network.ServerObject
	messageInterceptorHandler map[string]Network.TransmitMessageInterceptor
}

func CreateComponentServer() *ComponentServer {
	server := &ComponentServer{}
	server.init()
	return server
}
func (server *ComponentServer) init() {
	server.log = golog.New("core.ComponentServer")
	server.gameServerIdToSession = map[int32]cellnet.Session{}
	server.sessionIdToGameServerId = map[int64]int32{}
	server.messageInterceptorHandler = map[string]Network.TransmitMessageInterceptor{}
	server.RegisterMessage("InternalMessage.RegisterGameServerMsg", server.onGameServerRegister)
}
func (server *ComponentServer) Start(remoteConfigUrl string, serverType int32, serverId int32, isSingleInstance bool) {
	server.isSingleInstance = isSingleInstance
	server.logicId = serverType
	server.instanceId = serverId

	// 初始配置
	server.initConfig(remoteConfigUrl)
	//监听game server 连接
	server.listenGameServer()

	//注册退出函数
	Network.RegisterExitCallback(server.onExitFunction)

	Network.Start()
}

func (server *ComponentServer) Post(f func()) {
	Network.Post(f)
}

//注册消息处理器
func (server *ComponentServer) RegisterMessage(messageName string, callback cellnet.EventCallback) {
	Network.RegisterMessage(messageName, callback)
}

//创建一个定时器
func (server *ComponentServer) CreateTicker(duration time.Duration, callback func(*timer.Loop)) {
	Network.CreateTicker(duration, callback)
}

//注册消息拦截器 在io层拦截 并停止消息传递
func (server *ComponentServer) RegisterMessageInterceptor(messageName string, handler Network.TransmitMessageInterceptor) {
	if _, ok := server.messageInterceptorHandler[messageName]; ok {
		server.log.Errorf("already register message interceptor ", messageName, handler)
		return
	}
	server.messageInterceptorHandler[messageName] = handler
}

// 入站(接收)的事件处理,切记，这里是在io线程中
func (server *ComponentServer) OnInboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
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
		msgMeta := cellnet.MessageMetaByID(msg.MessageId)
		if msgMeta == nil {
			return inputEvent
		}
		obj := reflect.New(msgMeta.Type).Interface()
		err := msgMeta.Codec.Decode(msg.MessageBody, obj)
		if err != nil {
			return inputEvent
		}
		ev := &CustomCellnetEvent{}
		ev.Ses = inputEvent.Session()
		ev.Msg = obj
		ev.PlayerId = msg.PlayerId
		return ev
	default: //先检查是不是需要转发的消息
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
	}

	msglog.WriteRecvLogger(server.log, "tcp", inputEvent.Session(), inputEvent.Message())
	return inputEvent
}

// 出站(发送)的事件处理，切记，这里是在io线程中
func (server *ComponentServer) OnOutboundEvent(inputEvent cellnet.Event) (output cellnet.Event) {
	handled, err := rpc.ResolveOutboundEvent(inputEvent)

	if err != nil {
		server.log.Errorln("rpc.ResolveOutboundEvent:", err)
		return nil
	}
	if handled {
		return inputEvent
	}

	switch msg := inputEvent.Message().(type) {
	case *InternalMessage.InternalMsg:
		if msg.MsgUnEncode == nil {
			return nil
		}
		msgType := reflect.TypeOf(msg.MsgUnEncode)
		if nil == msgType {
			// error on get msg name
			server.log.Errorf("error on get message name")
			return nil
		}
		meta := cellnet.MessageMetaByType(msgType)
		if nil == meta {
			//
			server.log.Errorf("not found message meta by name", msg)
			return nil
		}

		msgInstance, err := meta.Codec.Encode(msg.MsgUnEncode, nil)
		if err != nil {
			server.log.Errorf("not found message meta by name", msg)
			return nil
		}
		msglog.WriteSendLogger(server.log, "tcp", inputEvent.Session(), msg.MsgUnEncode)
		msg.MsgUnEncode = nil
		msg.MessageBody = msgInstance.([]byte)
		msg.MessageId = meta.ID

		//需要先判断 是不是 rpc 消息
		if rpcevent, ok := inputEvent.(*rpc.RecvMsgEvent); ok {
			rpcevent.Msg = msg
			return rpcevent
		}

		return &cellnet.SendMsgEvent{Ses: inputEvent.Session(), Msg: msg}
	default:
		msglog.WriteSendLogger(server.log, "tcp", inputEvent.Session(), inputEvent.Message())
		return inputEvent
	}
	return inputEvent
}

func (server *ComponentServer) initConfig(remoteConfigUrl string) {

	// fetch config
	server.config = Config.FetchRemoteConfig(remoteConfigUrl)

	if server.isSingleInstance {
		for _, compServer := range server.config.SingletonComponentServerList {
			if compServer.LogicId == server.logicId {
				server.selfConfig = compServer
				break
			}
		}
	} else {
		for _, compServer := range server.config.MultiComponentServerList {
			if compServer.LogicId == server.logicId && compServer.Id == server.instanceId {
				server.selfConfig = compServer
				break
			}
		}
	}
}

func (server *ComponentServer) onExitFunction() {

	//停止game server监听
	if nil != server.listenGameServerSession {
		server.listenGameServerSession.Stop()
		server.listenGameServerSession = nil
	}
}
