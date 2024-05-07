package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	cfg "objectstore-server/service/upload/config"
	upProto "objectstore-server/service/upload/proto"
	"objectstore-server/service/upload/route"
	upRpc "objectstore-server/service/upload/rpc"
	"time"
)

func startRpcService() {
	registry := consul.NewRegistry() //a default to using env vars for master API

	service := micro.NewService(
		micro.Name("go.micro.service.upload"),
		micro.RegisterTTL(time.Second*10), // 声明超时时间, 避免consul不主动删掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second*5),
		micro.Registry(registry),
	)

	service.Init()

	upProto.RegisterUploadServiceHandler(service.Server(), new(upRpc.Upload))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startApiService() {
	router := route.Router()
	router.Run(cfg.UploadServiceHost)
}

func main() {
	// api 服务
	go startApiService()

	// rpc 服务
	startRpcService()
}
