# 服务端&客户端
- 服务端:account,dbproxy,download,upload,transfer
- 客户端:account,dbproxy,download,upload
- 后端:apigw

# 日志logs
- 日志地址:service/logs

# 上传文件upload
- 端口:28080                           (配置service/upload/config/config.go)
- 文件存放地址:"./data/fileserver"       (配置config/store.go)


# 下载文件download
- 端口:38080                           (配置"service/download/config/config.go")

# Mysql
- 账号:root
- 密码:111111
- 数据库:fileserver                    (配置"service/dbproxy/config/db.go")

