package cache

import "time"

const (
	MaxClientIDKey  = "max_client_id_{%d}_%d"
	LastMsgKey      = "last_msg_{%d}_%d"
	LoginSlotSetKey = "login_slot_set_{%d}" // 通过 hash tag保证在cluster模式上，这三个key都在一个shared分片实例上，减少网络调用
	TTL7D           = 7 * 24 * time.Hour
)
