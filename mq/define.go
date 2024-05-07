package mq

import (
	"objectstore-server/common"
)

// TransferData rabbitmq消息载体
type TransferData struct {
	FileHash      string
	CurLocation   string
	DestLocation  string
	DestStoreType common.StoreType
}
