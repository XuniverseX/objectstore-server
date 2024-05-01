package handler

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
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
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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

	w.Write(util.NewRespMsg(0, "OK", info).JSONBytes())
}

// UploadPartHandler 上传文件分块接口
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 获得文件句柄，用于存储分块内容
	fpath := "/Users/xuni/tmp/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 更新redis缓存状态
	conn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uploadId := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 通过uploadId查询redis并判断所有分块都上传完成
	data, err := redis.Values(conn.Do("HGETALL", "MP_"+uploadId))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed.", nil).JSONBytes())
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
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	// todo 分块合并
	// 更新文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(fileHash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, fileHash, filename, int64(fsize))

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
	// 响应
}
