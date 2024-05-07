package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"time"

	cfg "objectstore-server/service/download/config"
	dlProto "objectstore-server/service/download/proto"
	"objectstore-server/service/download/route"
	dlRpc "objectstore-server/service/download/rpc"
)

func startRpcService() {
	registry := consul.NewRegistry() //a default to using env vars for master API
	service := micro.NewService(
		micro.Name("go.micro.service.download"), // 在注册中心中的服务名称
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*5),
		micro.Registry(registry),
	)
	service.Init()

	dlProto.RegisterDownloadServiceHandler(service.Server(), new(dlRpc.Download))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startApiService() {
	router := route.Router()
	router.Run(cfg.DownloadServiceHost)
}

func main() {
	// api 服务
	go startApiService()

	// rpc 服务
	startRpcService()
}
