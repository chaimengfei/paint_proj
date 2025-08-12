package pkg

import (
	"cmf/paint_proj/configs"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"strings"
	"time"
)

func UploadImage(c *gin.Context) (string, error) {
	// 读取文件
	file, err := c.FormFile("file")
	if err != nil {
		return "", errors.New("获取文件失败:" + err.Error())
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开文件失败:" + err.Error())
	}
	defer src.Close()

	// OSS 配置
	endpoint := configs.Cfg.Oss.Endpoint
	accessKeyId := configs.Cfg.Oss.AccessKeyID
	accessKeySecret := configs.Cfg.Oss.AccessKeySecret
	bucketName := configs.Cfg.Oss.BucketName
	endpointSuffix := strings.TrimPrefix(endpoint, "https://")

	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		return "", errors.New("连接OSS失败:" + err.Error())
	}

	// 获取 bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", errors.New("获取Bucket失败:" + err.Error())
	}

	// 构造 OSS 文件路径
	filename := fmt.Sprintf("uploads/%d%s", time.Now().Unix(), filepath.Ext(file.Filename))

	// 上传到 OSS
	err = bucket.PutObject(filename, src)
	if err != nil {
		return "", errors.New("上传失败:" + err.Error())
	}
	// 返回图片 URL

	fileURL := fmt.Sprintf("https://%s.%s/%s", bucketName, endpointSuffix, filename)
	return fileURL, nil
}
