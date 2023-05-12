package util

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func CopyDirToS3(heapdumpName string, srcDir string, s3Bucket string, removeAfterCopy bool) error {
	// aws s3 cp my-local-folder s3://my-s3-bucket/ --recursive
	log.Printf("INFO|copyDirToS3(%s, %s)|exec aws s3 cp %s s3://%s/%s", srcDir, s3Bucket, srcDir, s3Bucket, heapdumpName)
	start := time.Now()
	cmd := exec.Command("aws", "s3", "cp", srcDir, fmt.Sprintf("s3://%s/%s", s3Bucket, heapdumpName), "--recursive")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	log.Printf("INFO|copyDirToS3(%s, %s)|AWS copy took %s", srcDir, s3Bucket, time.Now().Sub(start))
	if err != nil {
		return fmt.Errorf("ERROR|copyDirToS3(%s, %s)|AWS copy failed: %s", srcDir, s3Bucket, err.Error())
	}

	if removeAfterCopy {
		os.RemoveAll(srcDir)
	}
	return nil
}

func CopyFileToS3(srcFile string, s3Dest string, removeAfterCopy bool) error {
	log.Printf("INFO|copyFileToS3()|exec aws s3 cp %s s3://%s/", srcFile, os.Getenv("AWS_S3_BUCKET"))
	start := time.Now()
	cmd := exec.Command("aws", "s3", "cp", srcFile, fmt.Sprintf("s3://%s/", os.Getenv("AWS_S3_BUCKET")))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	log.Printf("INFO|copyFileToS3(%s, %s)|AWS copy took %s", srcFile, s3Dest, time.Now().Sub(start))
	if err != nil {
		return fmt.Errorf("ERROR|copyFileToS3(%s, %s)|AWS copy failed: %s", srcFile, s3Dest, err.Error())
	}

	if removeAfterCopy {
		os.Remove(srcFile)
	}

	return nil
}
