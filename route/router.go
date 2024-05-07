package route

import (
	"github.com/gin-gonic/gin"
	"objectstore-server/handler"
)

func Router() *gin.Engine {
	router := gin.Default()

	// 处理静态资源
	router.Static("/static/", "./static")

	// 不需要校验token的接口
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	router.GET("/user/signin", handler.SigninHandler)
	router.POST("/user/signin", handler.DoSigninHandler)

	// 加入拦截器中间件，后续路由都会经过拦截器
	router.Use(handler.HTTPInterceptor())

	// 文件存取接口
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/upload/suc", handler.UploadSuccessHandler)
	router.POST("/file/meta", handler.GetFileMetaHandler)
	router.POST("/file/query", handler.FileQueryHandler)
	router.GET("/file/download", handler.FileDownloadHandler)
	router.POST("/file/update", handler.FileMetaUpdateHandler)
	router.POST("/file/delete", handler.FileDeleteHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)

	// 秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	// 分片上传接口
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)

	// 用户相关接口
	router.POST("/user/info", handler.UserInfoHandler)

	return router
}
