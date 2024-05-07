package api

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"log"
	"net/http"
	m "objectstore-server/store/minio"
	"strings"

	"github.com/gin-gonic/gin"

	"objectstore-server/common"
	cfg "objectstore-server/config"
	dbcli "objectstore-server/service/dbproxy/client"
	"objectstore-server/store/oss"
)

// DownloadURLHandler 生成文件的下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	dbResp, err := dbcli.GetFileMeta(filehash)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": common.StatusServerError,
				"msg":  "server error",
			})
		return
	}

	// 转化为相应文件对象
	tblFile := dbcli.ToTableFile(dbResp.Data)

	// TODO: 判断文件存在OSS，还是Ceph，还是在本地
	if strings.HasPrefix(tblFile.FileAddr.String, cfg.TempLocalRootDir) ||
		strings.HasPrefix(tblFile.FileAddr.String, cfg.MinioRootDir) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "application/octet-stream", []byte(tmpURL))
	} else if strings.HasPrefix(tblFile.FileAddr.String, cfg.OSSRootDir) {
		// oss下载url
		signedURL := oss.DownloadURL(tblFile.FileAddr.String)
		log.Println(tblFile.FileAddr.String)
		c.Data(http.StatusOK, "application/octet-stream", []byte(signedURL))
	}
}

// DownloadHandler 文件下载接口
func DownloadHandler(c *gin.Context) {
	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	// TODO: 处理异常情况
	fResp, ferr := dbcli.GetFileMeta(fsha1)
	ufResp, uferr := dbcli.QueryUserFileMeta(username, fsha1)
	if ferr != nil || uferr != nil || !fResp.Suc || !ufResp.Suc {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": common.StatusServerError,
				"msg":  "server error",
			})
		return
	}
	uniqFile := dbcli.ToTableFile(fResp.Data)
	userFile := dbcli.ToTableUserFile(ufResp.Data)

	if strings.HasPrefix(uniqFile.FileAddr.String, cfg.TempLocalRootDir) {
		// 本地文件， 直接下载
		c.FileAttachment(uniqFile.FileAddr.String, userFile.FileName)
	} else if strings.HasPrefix(uniqFile.FileAddr.String, cfg.MinioRootDir) {
		// minio中的文件，通过minio api先下载
		object, err := m.Client().GetObject(context.Background(), "userfile",
			uniqFile.FileAddr.String, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			return
		}
		defer object.Close()

		data, err := io.ReadAll(object)
		if err != nil {
			log.Println(err)
			return
		}

		c.Header("content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
		c.Data(http.StatusOK, "application/octect-stream", data)
	}
}
