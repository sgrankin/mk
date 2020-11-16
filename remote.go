package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// for files that are http(s) urls, use the Last-Modified header
func updateHttpTimestamp(u *node) {
	resp, err := http.Head(u.name)
	if err != nil {
		log.Fatal(err)
	}
	lastModified := resp.Header.Get("Last-Modified")
	if lastModified == "" {
		// no Last-Modified header so lets assume that it
		// doesn't exist
		u.t = time.Unix(0, 0)
		u.exists = false
	} else {
		tmptime, err := time.Parse(time.RFC1123, lastModified)
		if err != nil {
			log.Fatal(err)
		}
		u.t = tmptime
		u.exists = true
	}
}

func updateS3Timestamp(u *node, uri *url.URL) {
	sess, err := session.NewSessionWithOptions(session.Options{
		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	})
	svc := s3.New(sess)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(uri.Host),
		Key:    aws.String(uri.Path[1:]),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Fatal(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatal(err.Error())
		}
		return
	}
	u.t = *result.LastModified
	u.exists = true
	//fmt.Println(result)
}
