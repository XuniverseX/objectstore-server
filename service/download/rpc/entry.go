package rpc

import (
	"context"
	cfg "objectstore-server/service/download/config"
	dlProto "objectstore-server/service/download/proto"
)

// Dwonload :download结构体
type Download struct{}

// DownloadEntry : 获取下载入口
func (u *Download) DownloadEntry(
	ctx context.Context,
	req *dlProto.DownloadReqEntry,
	res *dlProto.DownloadRespEntry) error {

	res.Entry = cfg.DownloadEntry
	return nil
}
