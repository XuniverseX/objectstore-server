package main

import (
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"objectstore-server/service/account/handler"
	"objectstore-server/service/account/proto"
	"time"
)

func main() {
	registry := consul.NewRegistry() //a default to using env vars for master API

	// 创建服务时指定注册中心
	service := micro.NewService(
		micro.Name("go.micro.service.account"),
		micro.RegisterTTL(time.Second*10), // 声明超时时间, 避免consul不主动删掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second*5),
		micro.Registry(registry),
	)

	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))

	if err := service.Run(); err != nil {
		panic(err)
	}
}
