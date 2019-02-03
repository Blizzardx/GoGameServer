package GameServer

import (
	"github.com/davyxu/cellnet"
	"github.com/Blizzardx/GoGameServer/Core/InternalMessage"
	"github.com/Blizzardx/GoGameServer/Core/Modules/Config"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"time"
)

func (server *GameServer) connectSingleComponentServer() {
	//链接单例组件服务器
	for _, serverInfo := range server.config.SingletonComponentServerList {
		serverObj := Network.ConnectToTCP(serverInfo.InternalAddress+":"+serverInfo.GameServerListenPort, server.onConnectComponentServerSessionChange, server, serverInfo)
		server.singleComponentServerMap[serverInfo.LogicId] = serverObj
	}
}

//主动连接： 组件服务器 连接发生变化
func (server *GameServer) onConnectComponentServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	serverConfig := arg.(*Config.NodeConfigComponentServerInfo)
	if eventName == "SessionConnected" {
		server.onComponentServerConnected(serverConfig.LogicId, ev.Session())

	} else if eventName == "SessionClosed" {
		server.onComponentServerDisConnect(serverConfig.LogicId)
	}
}
func (server *GameServer) onComponentServerConnected(serverType int32, session cellnet.Session) {
	if serverObj, ok := server.singleComponentServerMap[serverType]; ok {
		serverObj.Session = session
		session.Send(&InternalMessage.RegisterGameServerMsg{GameServerLogicId: server.logicId})
		server.log.Infoln("与单实例组件服务器 建立 链接", serverType)
	} else {
		server.log.Errorf("component server not connect")
	}
}
func (server *GameServer) onComponentServerDisConnect(serverType int32) {
	if serverObj, ok := server.singleComponentServerMap[serverType]; ok {
		serverObj.Session = nil
		server.log.Infoln("与单实例组件服务器 断开 链接", serverType)
	}
}

//给单例组件服务器 发送消息
func (server *GameServer) SendMessageToSingleComponentServer(serverType int32, msg interface{}) {
	if session, ok := server.singleComponentServerMap[serverType]; ok {
		if session.Session == nil {
			return
		}
		//
		session.Session.Send(msg)
	}
}

//给单例组件服务器 发送异步RPC 消息
func (server *GameServer) SendRPCMessageToSingleComponentServer(serverType int32, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	if session, ok := server.singleComponentServerMap[serverType]; ok {
		if session.Session == nil {
			return
		}
		//
		Network.RPC(session.Session, msg, timeOut, userCallback)
	}
}
