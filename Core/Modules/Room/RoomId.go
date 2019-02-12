package Room

import (
	"math/rand"
	"time"
)

//根据room id 解析server id
func ParserServerIdByRoomId(roomId int32) int32 {
	serverId := (roomId / 100000) % 10
	return serverId
}

//在现有的roomid上 添加server id 信息
func AttachServerIdToRoomId(roomId int32, serverId int32) int32 {
	roomId = roomId + serverId*100000
	return roomId
}

//根据match id 解析server id
func ParserServerIdByMatchId(matchId int32, battleServerCount int32) int32 {
	serverId := matchId % battleServerCount
	return serverId
}

func genRoomIds() {
	roomCount := 100000
	serverId := currentServerId
	//生成一个乱序的数组
	data := rand.New(rand.NewSource(time.Now().UnixNano())).Perm(roomCount)
	//把数组压到双向链表中，提高运行时的速度
	for i := 0; i < roomCount; i++ {
		roomIds.PushBack(AttachServerIdToRoomId(int32(data[i]), serverId))
	}
	log.Debugln("finished gen room id ", roomCount)
}

///如果为-1代表已经没有房间号可以使用了
func getRoomId() int32 {
	elem := roomIds.Front()
	if elem == nil {
		return -1
	}
	return int32(roomIds.Remove(elem).(int32))
}

//归还roomId
func recycleRoomId(roomId int32) {
	roomIds.PushBack(roomId)
}
