package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var config struct {
	Region      string `default:"us-west-1"`
	Bucket      string `required:"true"`
	FilePrefix  string `split_words:"true" default:"flood"`
	FileCount   int    `split_words:"true" default:"1100"`
	Concurrency int    `default:"10"`
}

func main() {
	envconfig.MustProcess("", &config)

	cctx, cancel := context.WithCancel(context.Background())

	sess := session.Must(session.NewSession(aws.NewConfig().WithRegion(config.Region)))

	uploader := s3manager.NewUploader(sess)

	var wg sync.WaitGroup

	fileRequests := make(chan int)

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go produceFiles(cctx, uploader, fileRequests, &wg)
	}

	go func() {
		for i := 0; i < config.FileCount; i++ {
			fileRequests <- i
		}
		close(fileRequests)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		s := <-sigChan
		logrus.Error("Received signal: ", s)
		cancel()
	}()

	wg.Wait()
}

func produceFiles(ctx context.Context, uploader *s3manager.Uploader, requests <-chan int, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case r, ok := <-requests:
			if !ok {
				wg.Done()
				return
			}

			fileName := fmt.Sprintf("%s-%d", config.FilePrefix, r)

			logrus.Infof("Uploading file %s", fileName)
			_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
				Bucket: aws.String(config.Bucket),
				Key:    aws.String(fileName),
				Body:   strings.NewReader(fileName),
			})
			if err != nil {
				logrus.Warnf("Upload of file '%s' failed: %v", fileName, err)
			}
		}
	}
}
