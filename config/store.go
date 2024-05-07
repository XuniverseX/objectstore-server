package config

import (
	cmn "objectstore-server/common"
)

const (
	// TempLocalRootDir 本地临时存储地址的路径
	TempLocalRootDir = "/Users/xuni/tmp/objectserver/"
	// TempPartRootDir 分块文件在本地临时存储地址的路径
	TempPartRootDir = "/Users/xuni/tmp/objectserver_part/"
	// MinioRootDir Minio的存储路径prefix
	MinioRootDir = "/minio"
	// OSSRootDir OSS的存储路径prefix
	OSSRootDir = "oss/"
	// CurrentStoreType 设置当前文件的存储类型
	CurrentStoreType = cmn.StoreAll
)
