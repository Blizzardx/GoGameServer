package OperationServer

import (
	"github.com/Blizzardx/GoGameServer/Core/Modules/Config"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"github.com/davyxu/cellnet"
	"time"
)

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
