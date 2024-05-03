package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	dblayer "objectstore-server/db"
	"objectstore-server/meta"
	"objectstore-server/store/oss"
	"objectstore-server/util"
	"os"
	"strconv"
	"time"
)

// UploadHandler GET获取上传页面；POST上传文件接口
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//返回上传的html页面
		file, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(file))
	} else if r.Method == http.MethodPost {
		//接受文件流及存储到本地目录
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Failed to get data,", err)
			return
		}
		defer file.Close()

		// 建立文件元信息
		fileMeta := meta.FileMeta{
			FileName:   header.Filename,
			Location:   "/Users/xuni/tmp/" + header.Filename,
			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("Failed to create file,", err)
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to save data into file,", err)
			return
		}

		//更新hash与fileMetas
		newFile.Seek(0, 0)
		fileMeta.FileHash = util.FileSha1(newFile)

		//同时将文件写入到minio
		newFile.Seek(0, 0)
		//stat, _ := newFile.Stat()
		//minioPath := "/minio/" + fileMeta.FileHash
		//m.Client().PutObject(context.Background(), "userfile", minioPath, newFile, stat.Size(),
		//	minio.PutObjectOptions{ContentType: "application/octet-stream"})
		//fileMeta.Location = minioPath

		ossPath := "oss/" + fileMeta.FileHash
		err = oss.GetBucket().PutObject(ossPath, newFile)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte("Upload failed!"))
			return
		}
		fileMeta.Location = ossPath

		//meta.UpdateFileMeta(fileMeta)
		meta.UpdateFileMetaToDB(fileMeta)

		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc, _ := dblayer.OnUserFileUploadFinished(username, fileMeta.FileHash,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}

	}
}

// UploadSuccessHandler 上传完成接口
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "上传成功！")
}

// GetFileMetaHandler 获取文件元信息接口
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fHash := r.Form.Get("filehash")
	//fMeta := meta.GetFileMeta(fHash)
	fMeta, err := meta.GetFileMetaFromDB(fHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// FileDownloadHandler 文件下载接口
func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fHash := r.Form.Get("filehash")
	fMeta, err := meta.GetFileMetaFromDB(fHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	file, err := os.Open(fMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application.octect-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fMeta.FileName+"\"")
	w.Write(data)
}

// DownloadURLHandler 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	// oss下载url
	signedURL := oss.GetDownloadUrl(row.FileAddr.String)
	w.Write([]byte(signedURL))

	// TODO: 判断文件存在OSS，还是Ceph，还是在本地
	//if strings.HasPrefix(row.FileAddr.String, "/tmp") {
	//	username := r.Form.Get("username")
	//	token := r.Form.Get("token")
	//	tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
	//		r.Host, filehash, username, token)
	//	w.Write([]byte(tmpUrl))
	//} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
	//	// TODO: ceph下载url
	//} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
	//	// oss下载url
	//	signedURL := oss.GetDownloadUrl(row.FileAddr.String)
	//	w.Write([]byte(signedURL))
	//}
}

// FileMetaUpdateHandler 文件重命名接口
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	op := r.Form.Get("opType")
	fHash := r.Form.Get("filehash")
	newFilename := r.Form.Get("filename")

	if op != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	curFilemeta := meta.GetFileMeta(fHash)
	curFilemeta.FileName = newFilename
	meta.UpdateFileMeta(curFilemeta)

	data, err := json.Marshal(curFilemeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler 文件删除接口
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fHash := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fHash)
	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fHash)

	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler 秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 从文件表中查询相同hash的记录
	fileMeta, err := meta.GetFileMetaFromDB(filehash)
	//if err != nil {
	//	fmt.Println(err)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}

	// 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 上传过则将文件信息写入用户文件表，返回成功
	suc, err := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
}
