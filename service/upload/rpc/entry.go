package rpc

import (
	"context"
	cfg "objectstore-server/service/upload/config"
	upProto "objectstore-server/service/upload/proto"
)

// Upload : upload结构体
type Upload struct{}

// UploadEntry : 获取上传入口
func (u *Upload) UploadEntry(
	ctx context.Context,
	req *upProto.UploadReqEntry,
	res *upProto.UploadRespEntry) error {

	res.Entry = cfg.UploadEntry
	return nil
}
