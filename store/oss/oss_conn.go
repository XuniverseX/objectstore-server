package oss

import (
	"fmt"
	cfg "objectstore-server/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

// Client 创建ossClient对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	c, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccesskeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ossCli = c
	return ossCli
}

// GetBucket 获取bucket
func GetBucket() *oss.Bucket {
	cli := Client()
	if cli == nil {
		return nil
	}
	bucket, err := cli.Bucket(cfg.OSSBucket)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bucket
}

// GetDownloadUrl 临时授权下载url
func GetDownloadUrl(objName string) string {
	signedURL, err := GetBucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return signedURL
}

// BuildLifecycleRule 针对指定bucket设置生命周期规则
func BuildLifecycleRule(bucketName string) {
	// 表示前缀为test的对象(文件)距最后修改时间30天后过期。
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	Client().SetBucketLifecycle(bucketName, rules)
}

// GenerateFileMeta 构造文件元信息
func GenerateFileMeta(metas map[string]string) []oss.Option {
	options := []oss.Option{}
	for k, v := range metas {
		options = append(options, oss.Meta(k, v))
	}
	return options
}
