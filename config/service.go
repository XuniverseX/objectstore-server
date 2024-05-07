package config

const (
	// UploadServiceHost 上传服务监听的地址
	UploadServiceHost = "0.0.0.0:8080"
	// UploadLBHost: 上传服务LB地址 (后续要更改问域名)
	UploadLBHost = "127.0.0.1:28080"
	// DownloadLBHost: 下载服务LB地址	(后续要更改为域名)
	DownloadLBHost = "127.0.0.1:38080"
	// TracerAgentHost tracing agent地址
	TracerAgentHost = "127.0.0.1:6831"
)
