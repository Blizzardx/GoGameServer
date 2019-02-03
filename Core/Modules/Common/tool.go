package Common

func ConvertPlayerIdToGameServerId(playerId int64, gameServerCount int) int32 {
	//gameserver id 必须 从 1 开始，递增，中间不可以有空位
	gsId := playerId % int64(gameServerCount)

	return int32(gsId) + 1
}
func ConvertPlayerStrIdToGameServerId(playerId string, gameServerCount int) int32 {
	intPlayerId := StringHash(playerId)
	//gameserver id 必须 从 1 开始，递增，中间不可以有空位
	gsId := intPlayerId % uint16(gameServerCount)

	return int32(gsId) + 1
}

// 字符串转为16位整形值
func StringHash(s string) (hash uint16) {

	for _, c := range s {

		ch := uint16(c)

		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}

	return
}
