package config

const (
	AsyncTransferEnable  = false // 是否开启文件异步转移(默认同步)
	RabbitURL            = "amqp://guest:guest@127.0.0.1:5672/"
	TransExchangeName    = "uploadserver.trans"
	TransOSSQueueName    = "uploadserver.trans.oss"
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	TransOSSRoutingKey   = "oss"
)
