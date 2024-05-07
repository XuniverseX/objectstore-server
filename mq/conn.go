package mq

import (
	"github.com/streadway/amqp"
	"log"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

func initChannel(url string) bool {
	if channel != nil {
		return true
	}

	c, err := amqp.Dial(url)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	conn = c

	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
