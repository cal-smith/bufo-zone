package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	dbhelper "bufo.zone"
	"bufo.zone/dbufo"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
)

const REPO_URL = "https://github.com/knobiknows/all-the-bufo/archive/refs/heads/main.zip"
const FILEPATH = "all_the_bufo.zip"

var shouldSyncToS3 = flag.Bool("to-s3", false, "sync bufos to s3")
var shouldSyncToDB = flag.Bool("to-db", false, "sync bufos to sqlite")

func download_bufos() {
	file, err := os.Create(FILEPATH)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Printf("downloading all the bufos from %s\n", REPO_URL)
	res, err := http.Get(REPO_URL)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		panic("non-200 status received")
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		panic(err)
	}
}

func list_bufos(service *s3.S3) []*s3.Object {
	var list []*s3.Object
	bucket := aws.String(os.Getenv("S3_BUFO_BUCKET"))
	res, err := service.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: bucket})
	for {
		if err != nil {
			panic(err)
		}
		list = append(list, res.Contents...)
		if !*res.IsTruncated {
			break
		}
		res, err = service.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            bucket,
			ContinuationToken: res.NextContinuationToken,
		})
	}

	return list
}

func sync_to_s3(service *s3.S3, uploader *s3manager.Uploader) {
	list := list_bufos(service)

	uploaded := 0
	skipped := 0
	bufos, err := zip.OpenReader(FILEPATH)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, bufo := range bufos.File {
		wg.Add(1)
		go func(bufo *zip.File) {
			defer wg.Done()
			if strings.Contains(bufo.Name, "/all-the-bufo/") {
				name := path.Base(bufo.Name)
				is_existing_bufo := slices.ContainsFunc(list, func(obj *s3.Object) bool {
					return *obj.Key == name
				})
				if !is_existing_bufo {
					fmt.Printf("uploading bufo %s\n", name)
					bufo_body, err := bufo.Open()
					if err != nil {
						fmt.Printf("error reading %s from zip\n", name)
						panic(err)
					}

					_, err = uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String(os.Getenv("S3_BUFO_BUCKET")),
						Key:    aws.String(name),
						Body:   bufo_body,
					})
					if err != nil {
						fmt.Printf("error uploading %s\n", name)
						panic(err)
					}
					uploaded += 1
				} else {
					fmt.Printf("skipping existing bufo %s\n", name)
					skipped += 1
				}
			} else {
				fmt.Printf("skipping %s since it is not a valid bufo\n", bufo.Name)
			}
		}(bufo)
	}
	wg.Wait()
	fmt.Printf("uploaded %d bufos and skipped %d bufos\n", uploaded, skipped)
}

func sync_to_sqlite(service *s3.S3) {
	ctx := context.Background()

	queries := dbhelper.GetDb(ctx)
	list := list_bufos(service)

	for _, bufo := range list {
		fmt.Printf("syncing: %s\n", *bufo.Key)
		queries.CreateBufo(ctx, dbufo.CreateBufoParams{Name: *bufo.Key, Created: time.Now()})
	}
	fmt.Printf("synced %d bufos\n", len(list))
}

func main() {
	flag.Parse()
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	if !*shouldSyncToDB && !*shouldSyncToS3 {
		fmt.Println("need to specify -to-s3 or -to-db (or both!)")
		os.Exit(0)
	}

	awssession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	s3svc := s3.New(awssession, aws.NewConfig().
		WithEndpoint(os.Getenv("S3_URL")).
		WithRegion("us-east-1").
		WithS3ForcePathStyle(true))

	if *shouldSyncToS3 {
		uploader := s3manager.NewUploaderWithClient(s3svc)
		download_bufos()
		sync_to_s3(s3svc, uploader)
	}

	if *shouldSyncToDB {
		sync_to_sqlite(s3svc)
	}
}
