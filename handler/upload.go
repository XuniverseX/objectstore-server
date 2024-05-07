package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"io"
	"log"
	"net/http"
	cmn "objectstore-server/common"
	cfg "objectstore-server/config"
	dblayer "objectstore-server/db"
	"objectstore-server/meta"
	"objectstore-server/mq"
	m "objectstore-server/store/minio"
	"objectstore-server/store/oss"
	"objectstore-server/util"
	"os"
	"strconv"
	"strings"
	"time"
)

func UploadHandler(c *gin.Context) {
	//返回上传的html页面
	data, err := os.ReadFile("./static/view/index.html")
	if err != nil {
		c.String(http.StatusBadGateway, "fhjdashfkjsa")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

// DoUploadHandler 处理文件上传
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "Upload failed",
			})
		}
	}()

	// 1. 从form表单中获得文件内容句柄
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("Failed to get form data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	// 2. 把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Printf("Failed to get file data, err:%s\n", err.Error())
		errCode = -2
		return
	}

	// 3. 构建文件元信息
	fileMeta := meta.FileMeta{
		FileName:   head.Filename,
		FileHash:   util.Sha1(buf.Bytes()), //　计算文件sha1
		FileSize:   int64(len(buf.Bytes())),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 4. 将文件写入临时存储位置
	fileMeta.Location = cfg.TempLocalRootDir + fileMeta.FileHash // 临时存储地址
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Printf("Failed to create file, err:%s\n", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil {
		fmt.Printf("Failed to save data into file, writtenSize:%d, err:%s\n", nByte, err)
		errCode = -4
		return
	}

	// 5. 同步或异步将文件转移到Minio/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部
	if cfg.CurrentStoreType == cmn.StoreMinio || cfg.CurrentStoreType == cmn.StoreAll {
		// 文件写入minio
		stat, _ := newFile.Stat()
		minioPath := "/minio/" + fileMeta.FileHash
		m.Client().PutObject(context.Background(), "userfile", minioPath, newFile, stat.Size(),
			minio.PutObjectOptions{ContentType: "application/octet-stream"})
		fileMeta.Location = minioPath
	}
	if cfg.CurrentStoreType == cmn.StoreOSS || cfg.CurrentStoreType == cmn.StoreAll {
		// 文件写入OSS存储
		ossPath := "oss/" + fileMeta.FileHash
		// 判断写入OSS为同步还是异步
		if !cfg.AsyncTransferEnable {
			// TODO: 设置oss中的文件名，方便指定文件名下载
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Println(err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		} else {
			// 写入异步转移任务队列
			data := mq.TransferData{
				FileHash:      fileMeta.FileHash,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: cmn.StoreOSS,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
			}
		}
	}
	//6.  更新文件表记录
	_ = meta.UpdateFileMetaToDB(fileMeta)

	// 7. 更新用户文件表
	username := c.Request.FormValue("username")
	fmt.Println("usernameeeeeeeee:", username)
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileHash,
		fileMeta.FileName, fileMeta.FileSize)
	if suc {
		c.Redirect(http.StatusFound, "/static/view/home.html")
	} else {
		errCode = -6
	}
}

// UploadSuccessHandler 上传完成接口
func UploadSuccessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Upload Finish",
	})
}

// GetFileMetaHandler 获取文件元信息接口
func GetFileMetaHandler(c *gin.Context) {
	fHash := c.Request.FormValue("filehash")
	//fMeta := meta.GetFileMeta(fHash)
	fMeta, err := meta.GetFileMetaFromDB(fHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  "Upload failed!",
		})
		return
	}

	if fMeta == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -4,
			"msg":  "No such file",
		})
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "Upload failed!",
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func FileQueryHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "Query failed!",
		})
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "Query failed!",
		})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// FileDownloadHandler 文件下载接口
func FileDownloadHandler(c *gin.Context) {
	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	// TODO: 处理异常情况
	fm, _ := meta.GetFileMetaFromDB(fsha1)
	userFile, _ := dblayer.QueryUserFileMeta(username, fsha1)

	if strings.HasPrefix(fm.Location, cfg.TempLocalRootDir) {
		// 本地文件直接下载
		c.FileAttachment(fm.Location, userFile.FileName)
	} else if strings.HasPrefix(fm.Location, cfg.MinioRootDir) {
		// minio中的文件，通过minio api先下载
		object, err := m.Client().GetObject(context.Background(), "userfile", fm.Location,
			minio.GetObjectOptions{})
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

		//bucket := ceph.GetCephBucket("userfile")
		//data, _ := bucket.Get(fm.Location)
		//	c.Header("content-type", "application/octect-stream")
		c.Header("content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
		c.Data(http.StatusOK, "application/octect-stream", data)
	}
}

// DownloadURLHandler 生成文件的下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	if strings.HasPrefix(row.FileAddr.String, cfg.TempLocalRootDir) ||
		strings.HasPrefix(row.FileAddr.String, cfg.MinioRootDir) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "octet-stream", []byte(tmpURL))
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		fmt.Println(row.FileAddr.String)
		c.Data(http.StatusOK, "octet-stream", []byte(signedURL))
	}
}

// FileMetaUpdateHandler 文件重命名接口
func FileMetaUpdateHandler(c *gin.Context) {
	op := c.Request.FormValue("opType")
	fHash := c.Request.FormValue("filehash")
	newFilename := c.Request.FormValue("filename")

	if op != "0" || len(newFilename) < 1 {
		c.Status(http.StatusForbidden)
		return
	}

	curFilemeta := meta.GetFileMeta(fHash)
	curFilemeta.FileName = newFilename
	meta.UpdateFileMeta(curFilemeta)

	data, err := json.Marshal(curFilemeta)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, data)
}

// FileDeleteHandler 文件删除接口
func FileDeleteHandler(c *gin.Context) {
	fHash := c.Request.FormValue("filehash")

	fMeta := meta.GetFileMeta(fHash)
	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fHash)

	c.Status(http.StatusOK)
}

// TryFastUploadHandler 秒传接口
func TryFastUploadHandler(c *gin.Context) {
	// 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 从文件表中查询相同hash的记录
	fileMeta, err := meta.GetFileMetaFromDB(filehash)
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.JSON(http.StatusOK, resp.JSONBytes())
		return
	}

	// 上传过则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.JSON(http.StatusOK, resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	c.JSON(http.StatusOK, resp.JSONBytes())
}

// UploadHandler GET获取上传页面；POST上传文件接口
//func UploadHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method == http.MethodGet {
//		//返回上传的html页面
//		file, err := os.ReadFile("./static/view/upload.html")
//		if err != nil {
//			io.WriteString(w, "internal server error")
//			return
//		}
//		io.WriteString(w, string(file))
//	} else if r.Method == http.MethodPost {
//		//接受文件流及存储到本地目录
//		file, header, err := r.FormFile("file")
//		if err != nil {
//			log.Println("Failed to get data,", err)
//			return
//		}
//		defer file.Close()
//
//		// 建立文件元信息
//		fileMeta := meta.FileMeta{
//			FileName:   header.Filename,
//			Location:   "/Users/xuni/tmp/" + header.Filename,
//			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
//		}
//
//		newFile, err := os.Create(fileMeta.Location)
//		if err != nil {
//			log.Println("Failed to create file,", err)
//			return
//		}
//		defer newFile.Close()
//
//		fileMeta.FileSize, err = io.Copy(newFile, file)
//		if err != nil {
//			log.Println("Failed to save data into file,", err)
//			return
//		}
//
//		//更新hash与fileMetas
//		newFile.Seek(0, 0)
//		fileMeta.FileHash = util.FileHash(newFile)
//
//		newFile.Seek(0, 0)
//
//		if cfg.CurrentStoreType == cmn.StoreMinio {
//			// 文件写入minio
//			stat, _ := newFile.Stat()
//			minioPath := "/minio/" + fileMeta.FileHash
//			m.Client().PutObject(context.Background(), "userfile", minioPath, newFile, stat.Size(),
//				minio.PutObjectOptions{ContentType: "application/octet-stream"})
//			fileMeta.Location = minioPath
//		} else if cfg.CurrentStoreType == cmn.StoreOSS {
//			// 文件写入OSS
//			//ossPath := "oss/" + fileMeta.FileHash
//			//err = oss.Bucket().PutObject(ossPath, newFile)
//			//if err != nil {
//			//	log.Println(err)
//			//	w.Write([]byte("Upload failed!"))
//			//	return
//			//}
//			//fileMeta.Location = ossPath
//
//			data := mq.TransferData{}
//			transData, _ := json.Marshal(data)
//			suc := mq.Publish(
//				cfg.TransExchangeName,
//				cfg.TransOSSRoutingKey,
//				transData)
//			if !suc {
//				// TODO 失败重试消息发送
//			}
//		}
//
//		//meta.UpdateFileMeta(fileMeta)
//		meta.UpdateFileMetaToDB(fileMeta)
//
//		// 更新用户文件表记录
//		r.ParseForm()
//		username := r.Form.Get("username")
//		suc, _ := dblayer.OnUserFileUploadFinished(username, fileMeta.FileHash,
//			fileMeta.FileName, fileMeta.FileSize)
//		if suc {
//			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
//		} else {
//			w.Write([]byte("Upload Failed"))
//		}
//
//	}
//}

//// UploadSuccessHandler 上传完成接口
//func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
//	io.WriteString(w, "上传成功！")
//}
//
//// GetFileMetaHandler 获取文件元信息接口
//func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//
//	fHash := r.Form.Get("filehash")
//	//fMeta := meta.GetFileMeta(fHash)
//	fMeta, err := meta.GetFileMetaFromDB(fHash)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	data, err := json.Marshal(fMeta)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	w.Write(data)
//}
//
//func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//
//	username := r.Form.Get("username")
//	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
//	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	data, err := json.Marshal(userFiles)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	w.Write(data)
//}
//
//// FileDownloadHandler 文件下载接口
//func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//
//	fHash := r.Form.Get("filehash")
//	fMeta, err := meta.GetFileMetaFromDB(fHash)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//	}
//
//	file, err := os.Open(fMeta.Location)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	defer file.Close()
//
//	data, err := io.ReadAll(file)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application.octect-stream")
//	w.Header().Set("Content-Disposition", "attachment;filename=\""+fMeta.FileName+"\"")
//	w.Write(data)
//}
//
//// DownloadURLHandler 生成文件的下载地址
//func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
//	filehash := r.Form.Get("filehash")
//	// 从文件表查找记录
//	row, _ := dblayer.GetFileMeta(filehash)
//
//	// oss下载url
//	signedURL := oss.DownloadURL(row.FileAddr.String)
//	w.Write([]byte(signedURL))
//
//	// TODO: 判断文件存在OSS，还是Ceph，还是在本地
//	//if strings.HasPrefix(row.FileAddr.String, "/tmp") {
//	//	username := r.Form.Get("username")
//	//	token := r.Form.Get("token")
//	//	tmpUrl := log.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
//	//		r.Host, filehash, username, token)
//	//	w.Write([]byte(tmpUrl))
//	//} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
//	//	// TODO: ceph下载url
//	//} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
//	//	// oss下载url
//	//	signedURL := oss.DownloadURL(row.FileAddr.String)
//	//	w.Write([]byte(signedURL))
//	//}
//}
//
//// FileMetaUpdateHandler 文件重命名接口
//func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//
//	op := r.Form.Get("opType")
//	fHash := r.Form.Get("filehash")
//	newFilename := r.Form.Get("filename")
//
//	if op != "0" {
//		w.WriteHeader(http.StatusForbidden)
//		return
//	}
//
//	curFilemeta := meta.GetFileMeta(fHash)
//	curFilemeta.FileName = newFilename
//	meta.UpdateFileMeta(curFilemeta)
//
//	data, err := json.Marshal(curFilemeta)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	w.Write(data)
//}
//
//// FileDeleteHandler 文件删除接口
//func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//	fHash := r.Form.Get("filehash")
//
//	fMeta := meta.GetFileMeta(fHash)
//	os.Remove(fMeta.Location)
//
//	meta.RemoveFileMeta(fHash)
//
//	w.WriteHeader(http.StatusOK)
//}
//
//// TryFastUploadHandler 秒传接口
//func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
//	r.ParseForm()
//
//	// 解析请求参数
//	username := r.Form.Get("username")
//	filehash := r.Form.Get("filehash")
//	filename := r.Form.Get("filename")
//	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
//	if err != nil {
//		log.Println(err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	// 从文件表中查询相同hash的记录
//	fileMeta, err := meta.GetFileMetaFromDB(filehash)
//	//if err != nil {
//	//	log.Println(err)
//	//	w.WriteHeader(http.StatusInternalServerError)
//	//	return
//	//}
//
//	// 查不到记录则返回秒传失败
//	if fileMeta == nil {
//		resp := util.RespMsg{
//			Code: -1,
//			Msg:  "秒传失败，请访问普通上传接口",
//		}
//		w.Write(resp.JSONBytes())
//		return
//	}
//
//	// 上传过则将文件信息写入用户文件表，返回成功
//	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
//	if suc {
//		resp := util.RespMsg{
//			Code: 0,
//			Msg:  "秒传成功",
//		}
//		w.Write(resp.JSONBytes())
//		return
//	}
//	resp := util.RespMsg{
//		Code: -2,
//		Msg:  "秒传失败，请稍后重试",
//	}
//	w.Write(resp.JSONBytes())
//}
