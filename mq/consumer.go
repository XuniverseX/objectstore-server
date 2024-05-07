package mq

import "log"

var done chan bool

// StartConsume 监听队列
func StartConsume(queueName, consumerName string, callback func(msg []byte) bool) {
	msgs, err := channel.Consume(queueName, consumerName,
		true,  //自动应答
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // false表示无消息会阻塞
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	done = make(chan bool)

	// 启动协程循环读取channel的数据
	go func() {
		for d := range msgs {
			suc := callback(d.Body)
			if suc {
				// TODO: 将任务写入错误队列，待后续处理
			}
		}
	}()

	// 没有信息过来则会一直阻塞，避免该函数退出
	<-done
	// 关闭通道
	channel.Close()
}

// StopConsume 停止监听队列
func StopConsume() {
	done <- true
}
