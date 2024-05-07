package mq

import (
	"github.com/streadway/amqp"
	"log"
	"objectstore-server/config"
)

// Publish 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	if !initChannel(config.RabbitURL) {
		return false
	}

	err := channel.Publish(exchange, routingKey,
		false, // 如果没有对应的queue, 是否保留该消息
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
