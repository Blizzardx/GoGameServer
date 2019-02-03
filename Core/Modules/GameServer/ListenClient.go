package GameServer

import (
	"github.com/davyxu/cellnet"
	"github.com/Blizzardx/GoGameServer/Core/InternalMessage"
	"github.com/Blizzardx/GoGameServer/Core/Network"
	"reflect"
	"runtime/debug"
)

func (server *GameServer) listenClient() bool {
	//
	if server.selfConfig.ClientProtocol == "ws" {
		//监听客户端连接
		server.listenClientSession = Network.ListenAtWS("http://0.0.0.0:"+server.selfConfig.ClientListenPort+"/", server.onListenClientSessionChange, server, nil)
		return true
	} else if server.selfConfig.ClientProtocol == "tcp" {
		//监听客户端连接
		server.listenClientSession = Network.ListenAtTCP("0.0.0.0:"+server.selfConfig.ClientListenPort, server.onListenClientSessionChange, server, nil)
		return true
	}

	server.log.Errorln("unknown client protocol ", server.selfConfig.ClientProtocol)
	return false
}

//监听： 有客户端 建立连接 或 有客户端离线
func (server *GameServer) onListenClientSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
	if eventName == "SessionConnected" {

	} else if eventName == "SessionClosed" {
		//清理session 和player 的映射关系,并 通知玩家处理器清理玩家缓存
		server.clearPlayerBySession(ev.Session().ID())
	}
}

//给客户端 发送内部消息
func (server *GameServer) sendInternalMessageToClient(msg *InternalMessage.InternalMsg) {
	if session, ok := server.playerIdToSession[msg.PlayerId]; ok {
		if session == nil {
			return
		}
		//内部容器里的消息需要decode 出来 发送的时候 才能正确encode
		meta := cellnet.MessageMetaByID(msg.MessageId)
		if nil == meta {
			//
			server.log.Errorf("not found message meta by id", msg.MessageId)
			return
		}
		//直接转发消息
		decodedMessage := reflect.New(meta.Type).Interface()
		if err := meta.Codec.Decode(msg.MessageBody, decodedMessage); err != nil {
			server.log.Errorf("error on decode msg by name", msg.MessageId)
			return
		}
		session.Send(decodedMessage)
	}
}

//通过session 获取player
func (server *GameServer) getPlayerIdBySession(sessionId int64) string {
	if playerId, ok := server.sessionIdToPlayerId[sessionId]; ok {
		return playerId
	}
	return ""
}

//连接被动断开，通知服务器
func (server *GameServer) clearPlayerBySession(sessionId int64) {

	playerId := server.getPlayerIdBySession(sessionId)
	delete(server.sessionIdToPlayerId, sessionId)
	if playerId == "" {
		return
	}
	//清理session 和player 的映射关系
	delete(server.playerIdToSession, playerId)
	//有玩家掉线，需要清理在线玩家
	server.onPlayerDisconnectHandler(playerId)
}

//主动断开与客户端的连接
func (server *GameServer) clearPlayerByPlayer(playerId string) {
	//清理session 和player 的映射关系
	if session, ok := server.playerIdToSession[playerId]; ok {
		delete(server.playerIdToSession, playerId)
		if session == nil {
			server.log.Errorf("session already closed ", playerId)
			return
		}
		delete(server.sessionIdToPlayerId, session.ID())
		session.Close()
	}
}

func (server *GameServer) OnPlayerLogin(playerId string, session cellnet.Session) {
	//判断是不是有玩家在线
	if session, ok := server.playerIdToSession[playerId]; ok {
		//有相同player id 的玩家 建立了连接，吧这个玩家清理，然后发送kick 消息
		if session != nil {
			server.onKickPlayerHandler(playerId)
		}
		server.clearPlayerByPlayer(playerId)
	}
	server.sessionIdToPlayerId[session.ID()] = playerId
	server.playerIdToSession[playerId] = session
}

func (server *GameServer) SendToClient(playerId string, msg interface{}) bool {
	if msg == nil {
		server.log.Errorln("can't send nil message", playerId, string(debug.Stack()))
		return false
	}
	if session, ok := server.playerIdToSession[playerId]; ok {
		if session == nil {
			return false
		}
		session.Send(msg)
		return true
	}
	return false
}
func (server *GameServer) GetPlayerIdBySession(sessionId int64) string {
	if playerId, ok := server.sessionIdToPlayerId[sessionId]; ok {
		return playerId
	}
	return ""
}
