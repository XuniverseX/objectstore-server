package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"objectstore-server/meta"
	"objectstore-server/util"
	"os"
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
		fileMeta.Hash = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		meta.UpdateFileMetaToDB(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)

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
