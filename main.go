package main

import (
	"fmt"
	"net/http"
	"objectstore-server/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSuccessHandler)
	http.HandleFunc("GET /file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.FileDownloadHandler)
	http.HandleFunc("POST /file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)

	http.HandleFunc("/user/signup", handler.SignupHandler)

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		fmt.Println("Failed to start server,", err)
	}
}
