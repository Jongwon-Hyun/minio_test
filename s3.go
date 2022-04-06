package s3

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
)

func UploadFile(file multipart.File, fileName string, env string) (*manager.UploadOutput, error) {
	s3Client, err := newS3Storage(env)
	if err != nil {
		return nil, err
	}
	uploader := manager.NewUploader(s3Client)

	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(getBucketName(env)),
		Key:    aws.String(fileName),
		Body:   file,
		// TODO minio kms 세팅이 복잡해서 시간날 때 구현해 볼 것
		//ServerSideEncryption: "aws:kms",
	})

	return result, err
}

func DownloadFile(objectKey string, env string) ([]byte, error) {
	s3Client, err := newS3Storage(env)
	if err != nil {
		return nil, err
	}

	buffer := manager.NewWriteAtBuffer([]byte{})
	downloader := manager.NewDownloader(s3Client)

	numBytes, err := downloader.Download(context.TODO(), buffer, &s3.GetObjectInput{
		Bucket: aws.String(getBucketName(env)),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}

	if numBytes < 1 {
		return nil, errors.New("zero bytes written to memory")
	}

	return buffer.Bytes(), nil
}

func newS3Storage(env string) (*s3.Client, error) {
	awsCfg, err := createAWSCfg(env)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return s3Client, nil
}

func createAWSCfg(env string) (aws.Config, error) {
	// TODO 하드코딩된 설정 값들을 환경변수에서 가져와서 세팅하게 변경해야 함
	switch env {
	case "dev":
		credential := credentials.NewStaticCredentialsProvider("accessKey", "secretKey", "")
		endPointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               "http://localhost:9000",
				SigningRegion:     "ap-northeast-2",
				HostnameImmutable: true,
			}, nil
		})
		awsCfg, err := awsConfig.LoadDefaultConfig(
			context.TODO(),
			awsConfig.WithCredentialsProvider(credential),
			awsConfig.WithEndpointResolverWithOptions(endPointResolver),
		)
		if err != nil {
			return aws.Config{}, err
		}
		return awsCfg, nil
	case "prod":
		credential := credentials.NewStaticCredentialsProvider("awsAccessKey", "awsSecretKey", "")
		awsCfg, err := awsConfig.LoadDefaultConfig(
			context.TODO(),
			awsConfig.WithCredentialsProvider(credential),
			awsConfig.WithRegion("ap-northeast-2"),
		)
		if err != nil {
			return aws.Config{}, err
		}
		return awsCfg, nil
	default:
		return aws.Config{}, errors.New("env is not defined")
	}
}

func getBucketName(env string) string {
	switch env {
	case "dev":
		return "dev-bucket"
	case "prod":
		return "prod-bucket"
	default:
		return "dev-bucket"
	}
}
