package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Get MinIO configuration from environment variables with fallbacks
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	accessKeyID := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	secretAccessKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatalln("Error with MinIO client creation:", err)
	}

	ctx := context.Background()
	bucketName := "test-bucket"
	location := "us-east-1"

	fmt.Println("Bucket creation...")
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			fmt.Printf("Bucket %s already exist\n", bucketName)
		} else {
			log.Fatalln("Error with bucket creation:", err)
		}
	} else {
		fmt.Printf("Bucket %s successfully created\n", bucketName)
	}

	fmt.Println("\nFile upload to bucket...")
	objectName := "test-file.txt"
	filePath := "test-file.txt"
	contentType := "text/plain"

	testData := "Test file for MinIO"
	err = os.WriteFile(filePath, []byte(testData), 0644)
	if err != nil {
		log.Fatalln("Error with creating test file:", err)
	}

	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Fatalln("Error with file uploading:", err)
	}
	fmt.Printf("File %s successfully uploaded the size is %d bytes\n", objectName, info.Size)

	fmt.Println("\nUploading string data to bucket")
	objectName2 := "string-data.txt"
	stringData := "Data uploaded from a string"
	reader := strings.NewReader(stringData)

	info2, err := minioClient.PutObject(ctx, bucketName, objectName2, reader, int64(len(stringData)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Fatalln("Error uploading string:", err)
	}
	fmt.Printf("String uploaded as %s, size: %d bytes\n", objectName2, info2.Size)

	fmt.Println("\nList of objects in bucket:")
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    "",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			log.Fatalln("Error getting object list:", object.Err)
		}
		fmt.Printf("- %s (size: %d, modified: %s)\n", object.Key, object.Size, object.LastModified)
	}

	fmt.Println("\nDownloading file...")
	downloadPath := "downloaded-" + objectName
	err = minioClient.FGetObject(ctx, bucketName, objectName, downloadPath, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln("Error downloading file:", err)
	}
	fmt.Printf("File downloaded as %s\n", downloadPath)

	fmt.Println("\nReading object as stream...")
	object, err := minioClient.GetObject(ctx, bucketName, objectName2, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln("Error getting object:", err)
	}
	defer object.Close()

	buf := make([]byte, 1024)
	n, err := object.Read(buf)
	if err != nil && err.Error() != "EOF" {
		log.Fatalln("Error reading object:", err)
	}
	fmt.Printf("Object content: %s\n", string(buf[:n]))

	fmt.Println("\nDeleting objects...")
	err = minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Fatalln("Error deleting object:", err)
	}
	fmt.Printf("Object %s deleted\n", objectName)

	err = minioClient.RemoveObject(ctx, bucketName, objectName2, minio.RemoveObjectOptions{})
	if err != nil {
		log.Fatalln("Error deleting object:", err)
	}
	fmt.Printf("Object %s deleted\n", objectName2)

	fmt.Println("\nDeleting bucket...")
	err = minioClient.RemoveBucket(ctx, bucketName)
	if err != nil {
		log.Fatalln("Error deleting bucket:", err)
	}
	fmt.Printf("Bucket %s deleted\n", bucketName)

	os.Remove(filePath)
	os.Remove(downloadPath)

	fmt.Println("\nAll operations completed successfully!")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}