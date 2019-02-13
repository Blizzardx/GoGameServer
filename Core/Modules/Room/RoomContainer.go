package Room

import (
	"github.com/Blizzardx/GoGameServer/Core/Common"
	"github.com/davyxu/cellnet"
	"time"
)

type roomMsg struct {
	msgBody  interface{}
	playerId string
}
type RoomContainer struct {
	roomId         int32
	createTime     time.Time
	roomLogic      IRoom
	maxSilenceTime time.Duration //最长静默时间 收不到消息 就算静默 超过静默时间 自动回收房间资源 删除房间
	tickRate       time.Duration
	msgQueue       *Common.SyncQueue
	isRunning      bool
	lastAliveTime  time.Time
}

func (room *RoomContainer) init(roomLogic IRoom, roomId int32) {
	room.roomId = roomId
	room.msgQueue = Common.NewSyncQueue()
	room.roomLogic = roomLogic
	room.maxSilenceTime = room.roomLogic.GetMaxSilenceTime()
	room.tickRate = room.roomLogic.GetTickRate()
	room.isRunning = true
	room.roomLogic.OnInit()
	room.createTime = time.Now()
	Common.SafeCall(room.beginContainer)
}
func (room *RoomContainer) delete() {
	room.isRunning = false
}

func (room *RoomContainer) beginContainer() {
	tick := time.NewTicker(room.tickRate * time.Millisecond)
	roomExpireCheckTicker := time.NewTicker(5 * time.Minute)
	for room.isRunning {
		select {
		case <-tick.C:
			room.handleRoomMsg()
			room.roomLogic.OnTick()
		case <-roomExpireCheckTicker.C:
			room.checkRoomExpire()
		}
	}
	room.onDelete()
}
func (room *RoomContainer) onDelete() {
	room.roomLogic.OnDelete()
}
func (room *RoomContainer) onRoomMessage(playerId string, msgBody interface{}) {
	room.msgQueue.Offer(&roomMsg{msgBody: msgBody, playerId: playerId})
	room.RefreshAliveTime()
}
func (room *RoomContainer) handleRoomMsg() {
	for {
		msgInfo := room.msgQueue.Poll()
		if msgInfo == nil {
			break
		}
		msg := msgInfo.(*roomMsg)
		if msg == nil {
			log.Errorln("parser roomMsg is nil")
			break
		}
		meta := cellnet.MessageMetaByMsg(msg.msgBody)
		if nil == meta {
			log.Errorln("parser roomMsg on error ,can't get message meta by msg", msg)
			break
		}
		msgName := meta.FullName()
		room.roomLogic.OnRoomMessage(msg.playerId, msgName, msg.msgBody)
	}
}
func (room *RoomContainer) checkRoomExpire() {
	if room.lastAliveTime.Add(room.maxSilenceTime).Before(time.Now()) {
		DeleteRoom(room.roomId)
	}
}

//----------------------public interface-----------------------------
func (room *RoomContainer) RefreshAliveTime() {
	room.lastAliveTime = time.Now()
}
func (room *RoomContainer) GetId() int32 {
	return room.roomId
}
func (room *RoomContainer) GetCreateTime() time.Time {
	return room.createTime
}
func (room *RoomContainer) GetLogic() IRoom {
	return room.roomLogic
}
