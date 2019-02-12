package Room

import (
	"container/list"
	"github.com/davyxu/golog"
)

//显示初始化 需要传当前server的server id
func Init(serverId int32) {
	currentServerId = serverId
	genRoomIds()
}

var (
	log             = golog.New("core.roomManager")
	roomIds         = list.New()
	currentServerId int32
	roomMap         = map[int32]*RoomContainer{}
)

//创建房间
func CreateRoom(roomLogic IRoom) *RoomContainer {
	if nil == roomLogic {
		return nil
	}
	roomId := getRoomId()
	if roomId == -1 {
		return nil
	}
	room := &RoomContainer{}
	room.init(roomLogic, roomId)

	roomMap[roomId] = room
	return room
}

//删除房间
func DeleteRoom(id int32) {
	if room, ok := roomMap[id]; ok {
		delete(roomMap, id)
		recycleRoomId(id)
		room.delete()
	}
}

//查找房间 通过id
func GetRoom(id int32) *RoomContainer {
	if room, ok := roomMap[id]; ok {
		return room
	}
	return nil
}

//房间消息
func OnRoomMessage(roomId int32, playerId string, msgBody interface{}) {
	room := GetRoom(roomId)
	if room == nil {
		log.Errorln("not found room by id on room msg ", roomId, playerId, msgBody)
		return
	}
	room.onRoomMessage(playerId, msgBody)
}
