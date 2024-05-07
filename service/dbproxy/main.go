package main

import (
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"log"
	dbProxy "objectstore-server/service/dbproxy/proto"
	dbRpc "objectstore-server/service/dbproxy/rpc"
	"time"
)

func startRpcService() {
	registry := consul.NewRegistry() //a default to using env vars for master API

	service := micro.NewService(
		micro.Name("go.micro.service.dbproxy"), // 在注册中心中的服务名称
		micro.RegisterTTL(time.Second*10),      // 声明超时时间, 避免consul不主动删掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second*5),
		micro.Registry(registry),
	)
	service.Init()

	dbProxy.RegisterDBProxyServiceHandler(service.Server(), new(dbRpc.DBProxy))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func main() {
	startRpcService()
	// res, err := mapper.FuncCall("/user/UserExist", []interface{}{"haha"}...)
	// log.Printf("error: %+v\n", err)
	// log.Printf("result: %+v\n", res[0].Interface())

	// res, err = mapper.FuncCall("/user/UserExist", []interface{}{"admin"}...)
	// log.Printf("error: %+v\n", err)
	// log.Printf("result: %+v\n", res[0].Interface())
}
