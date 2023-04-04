package main

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
)

func Handler(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		s3Bucket := record.S3.Bucket.Name
		s3ObjectKey := record.S3.Object.Key

		// Create a new session with AWS SDK
		sess := session.Must(session.NewSession())

		// Create an S3 service client
		svc := s3.New(sess)

		// Download the file from S3
		downloadedFile, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3ObjectKey),
		})
		if err != nil {
			return err
		}
		defer downloadedFile.Body.Close()

		// Decode the image
		img, _, err := image.Decode(downloadedFile.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Resize the image
		resizedImg := imaging.Resize(img, 300, 0, imaging.Lanczos)

		// Encode the resized image as JPEG
		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, resizedImg, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Upload the resized file back to S3
		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(s3Bucket + "-resized"),
			Key:    aws.String(s3ObjectKey),
			Body:   bytes.NewReader(buf.Bytes()),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
