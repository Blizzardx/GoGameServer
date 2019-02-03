package ComponentServer

import (
	"github.com/davyxu/cellnet"
	"litgame.cn/Server/Core/InternalMessage"
	"litgame.cn/Server/Core/Modules/Common"
	"litgame.cn/Server/Core/Network"
	"time"
)

func (server *ComponentServer) listenGameServer() {
	//监听其他gameServer服务器连接
	server.listenGameServerSession = Network.ListenAtTCP("0.0.0.0:"+server.selfConfig.GameServerListenPort, server.onListenServerSessionChange, server, nil)
}

//监听： 有GameServer 建立连接 或 有GameServer离线
func (server *ComponentServer) onListenServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	if eventName == "SessionConnected" {

	} else if eventName == "SessionClosed" {
		//清理session game server 的映射关系
		if gsId, ok := server.sessionIdToGameServerId[ev.Session().ID()]; ok {
			delete(server.gameServerIdToSession, gsId)
			delete(server.sessionIdToGameServerId, ev.Session().ID())
			server.log.Infoln("有GameServer离线", gsId)
		}
	}
}

//监听： 有GameServer 建立连接
func (server *ComponentServer) onGameServerRegister(ev cellnet.Event) {
	msg := ev.Message().(*InternalMessage.RegisterGameServerMsg)
	server.log.Infoln("有GameServer 建立连接", msg.GameServerLogicId)

	server.gameServerIdToSession[msg.GameServerLogicId] = ev.Session()
	server.sessionIdToGameServerId[ev.Session().ID()] = msg.GameServerLogicId
}

//发送消息给 playerId 对应的gameServer
func (server *ComponentServer) SendToGameServer(playerId string, msg interface{}) {
	gsId := Common.ConvertPlayerStrIdToGameServerId(playerId, len(server.config.GameServerList))
	server.sendToGameServer(gsId, msg)
}

//给单例组件服务器 发送异步RPC 消息
func (server *ComponentServer) SendRPCMessageToGameServer(playerId string, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	gsId := Common.ConvertPlayerStrIdToGameServerId(playerId, len(server.config.GameServerList))
	if session, ok := server.gameServerIdToSession[gsId]; ok {
		Network.RPC(session, msg, timeOut, userCallback)
	}
}

//发送消息直接给 player，先把消息包装成内部消息，然后通过对应的 gameServer 转发给player
func (server *ComponentServer) SendToPlayer(playerId string, msg interface{}) {
	gsId := Common.ConvertPlayerStrIdToGameServerId(playerId, len(server.config.GameServerList))

	internalMsg := &InternalMessage.InternalMsg{PlayerId: playerId, MessageId: 0, MsgUnEncode: msg}
	server.sendToGameServer(gsId, internalMsg)
}

func (server *ComponentServer) sendToGameServer(gsId int32, msg interface{}) {
	if session, ok := server.gameServerIdToSession[gsId]; ok {
		session.Send(msg)
	}
}
