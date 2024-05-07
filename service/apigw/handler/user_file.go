package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	userProto "objectstore-server/service/account/proto"
)

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(c *gin.Context) {
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")

	rpcResp, err := userCli.UserFiles(context.TODO(), &userProto.ReqUserFile{
		Username: username,
		Limit:    int32(limitCnt),
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	newFileName := c.Request.FormValue("filename")

	log.Println("fileSha:", fileSha1)
	if opType != "0" || len(newFileName) < 1 {
		c.Status(http.StatusForbidden)
		return
	}

	rpcResp, err := userCli.UserFileRename(context.TODO(), &userProto.ReqUserFileRename{
		Username:    username,
		Filehash:    fileSha1,
		NewFileName: newFileName,
	})

	log.Println("NewFileName:", newFileName)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}
	// 返回数据给客户端
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}
