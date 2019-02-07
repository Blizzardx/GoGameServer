package OperationServer

import (
	"github.com/Blizzardx/GoGameServer/Core/Modules/Config"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"github.com/davyxu/cellnet"
	"time"
)

func (server *operationServer) connectMultiComponentServer() {

	//链接多实例组件服务器
	for _, serverInfo := range server.config.MultiComponentServerList {
		serverObj := Network.ConnectToTCP(serverInfo.InternalAddress+":"+serverInfo.GMListenPort, server.onConnectMultiComponentServerSessionChange, server, serverInfo)
		if serverMap, ok := server.multiComponentServerMap[serverInfo.LogicId]; ok {
			serverMap[serverInfo.Id] = serverObj
		} else {
			server.multiComponentServerMap[serverInfo.LogicId] = map[int32]*Network.ServerObject{}
			server.multiComponentServerMap[serverInfo.LogicId][serverInfo.Id] = serverObj
		}
	}
}

//主动连接： 多实例组件服务器 连接发生变化
func (server *operationServer) onConnectMultiComponentServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	serverConfig := arg.(*Config.NodeConfigComponentServerInfo)
	if eventName == "SessionConnected" {
		server.onMultiComponentServerConnected(serverConfig.LogicId, serverConfig.Id, ev.Session())

	} else if eventName == "SessionClosed" {
		server.onMultiComponentServerDisConnect(serverConfig.LogicId, serverConfig.Id)
	}
}
func (server *operationServer) onMultiComponentServerConnected(serverType int32, serverId int32, session cellnet.Session) {
	if serverMap, ok := server.multiComponentServerMap[serverType]; ok {
		if serverObj, ok := serverMap[serverId]; ok {
			serverObj.Session = session
			server.log.Infoln("与多实例组件服务器 建立 链接", serverType, serverId)
		} else {
			server.log.Errorf("multi component server not connect")
		}
	} else {
		server.log.Errorf("multi component server not connect")
	}
}
func (server *operationServer) onMultiComponentServerDisConnect(serverType int32, serverId int32) {
	if serverMap, ok := server.multiComponentServerMap[serverType]; ok {
		if serverObj, ok := serverMap[serverId]; ok {
			serverObj.Session = nil
			server.log.Infoln("与多实例组件服务器 断开 链接", serverType, serverId)
		}
	}
}

//给单例组件服务器 发送消息
func (server *operationServer) SendMessageToMultiComponentServer(serverType int32, serverId int32, msg interface{}) {
	if serverMap, ok := server.multiComponentServerMap[serverType]; ok {
		if session, ok := serverMap[serverId]; ok {
			if session.Session == nil {
				return
			}
			//
			session.Session.Send(msg)
		}
	}
}

//给单例组件服务器 发送异步RPC 消息
func (server *operationServer) SendRPCMessageToMultiComponentServer(serverType int32, serverId int32, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	if serverMap, ok := server.multiComponentServerMap[serverType]; ok {
		if session, ok := serverMap[serverId]; ok {
			if session.Session == nil {
				return
			}
			//
			Network.RPC(session.Session, msg, timeOut, userCallback)
		}
	}
}
