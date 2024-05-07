package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"math"
	"net/http"
	rPool "objectstore-server/cache/redis"
	dblayer "objectstore-server/db"
	"objectstore-server/util"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	chunk_size = 5 * 1024 * 1024
)

// MultipartUploadInfo 切片信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int64
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitialMultipartUploadHandler 初始化分块上传接口
func InitialMultipartUploadHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params invalid",
		})
		return
	}

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 生成分块上传的初始化信息
	info := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   int64(filesize),
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  chunk_size,
		ChunkCount: int(math.Ceil(float64(filesize) / chunk_size)),
	}

	// 将初始化信息写入redis缓存
	conn.Do("HSET", "MP_"+info.UploadID, "chunkcount", info.ChunkCount)
	conn.Do("HSET", "MP_"+info.UploadID, "filehash", info.FileHash)
	conn.Do("HSET", "MP_"+info.UploadID, "filesize", info.FileSize)

	resp := util.NewRespMsg(0, "OK", info)
	c.Data(http.StatusOK, "", resp.JSONBytes())
}

// UploadPartHandler 上传文件分块接口
func UploadPartHandler(c *gin.Context) {
	//	username := r.Form.Get("username")
	uploadID := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 获得文件句柄，用于存储分块内容
	fpath := "/Users/xuni/tmp/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "Upload part failed",
			"data": nil,
		})
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 更新redis缓存状态
	conn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	c.JSON(
		http.StatusOK, gin.H{
			"code": 0,
			"msg":  "OK",
			"data": nil,
		})
}

// CompleteUploadHandler 通知上传合并
func CompleteUploadHandler(c *gin.Context) {
	uploadId := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	fileHash := c.Request.FormValue("filehash")
	filesize := c.Request.FormValue("filesize")
	filename := c.Request.FormValue("filename")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 通过uploadId查询redis并判断所有分块都上传完成
	data, err := redis.Values(conn.Do("HGETALL", "MP_"+uploadId))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Complete upload failed.",
			"data": nil,
		})
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
			continue
		}
		if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "invalid request",
			"data": nil,
		})
		return
	}
	// todo 分块合并

	// 更新文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(fileHash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, fileHash, filename, int64(fsize))

	// 6. 响应处理结果
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "OK",
		"data": nil,
	})
	// 响应
}
