package main

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	m "objectstore-server/store/minio"
	"os"
)

func main() {
	bucketName := "testbucket1"
	objectName := "testobject333"
	cli := m.Client()
	ctx := context.Background()

	//err := cli.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	//if err != nil {
	//	fmt.Println(err)
	//}

	objInfo, err := cli.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", objInfo.Key)

	file, err := os.Open("/Users/xuni/Downloads/Hearthstone-Setup.zip")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	//UserMeta := map[string]string{
	//	"origin_name": "Hearthstone-Setup.zip",
	//}
	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
	}
	size := fileStat.Size()
	uploadInfo, err := cli.PutObject(ctx, bucketName, objectName, file, size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully uploaded bytes:", uploadInfo)

	objInfo1, err := cli.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", objInfo1.Key)
}
