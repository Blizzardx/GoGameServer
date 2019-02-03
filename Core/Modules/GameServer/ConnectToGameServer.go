package GameServer

import (
	"github.com/davyxu/cellnet"
	"litgame.cn/Server/Core/Modules/Config"
	"litgame.cn/Server/Core/Network"
	"time"
)

func (server *GameServer) connectGameServer() {
	for _, serverInfo := range server.config.GameServerList {
		if serverInfo.LogicId == server.logicId {
			continue
		}
		serverObj := Network.ConnectToTCP(serverInfo.InternalAddress+":"+serverInfo.GameServerListenPort, server.onConnectServerSessionChange, server, serverInfo)
		server.gameServerMap[serverInfo.LogicId] = serverObj
	}
}

//主动连接： GameServer 连接发生变化
func (server *GameServer) onConnectServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	serverConfig := arg.(*Config.NodeConfigGameServerInfo)
	if eventName == "SessionConnected" {
		server.onGameServerServerConnected(serverConfig.LogicId, ev.Session())

	} else if eventName == "SessionClosed" {
		server.onGameServerServerDisConnect(serverConfig.LogicId)
	}
}

func (server *GameServer) onGameServerServerConnected(serverType int32, session cellnet.Session) {
	if serverObj, ok := server.gameServerMap[serverType]; ok {
		serverObj.Session = session
		server.log.Infoln("与game server 服务器 建立 链接", serverType)
	}
}
func (server *GameServer) onGameServerServerDisConnect(serverType int32) {
	if serverObj, ok := server.gameServerMap[serverType]; ok {
		serverObj.Session = nil
		server.log.Infoln("与game server 服务器 断开 链接", serverType)
	}
}

//给GameServer 发送消息
func (server *GameServer) sendMessageToGameServer(gsId int32, msg interface{}) {
	if session, ok := server.gameServerMap[gsId]; ok {
		if session.Session == nil {
			return
		}
		//
		session.Session.Send(msg)
	}
}

//给GameServer 发送异步RPC 消息
func (server *GameServer) sendRPCMessageToGameServer(gsId int32, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	if session, ok := server.gameServerMap[gsId]; ok {
		if session.Session == nil {
			return
		}
		//
		Network.RPC(session.Session, msg, timeOut, userCallback)
	}
}
