package main

// Reference
//   https://aws.amazon.com/blogs/developer/mocking-out-then-aws-sdk-for-go-for-unit-testing/

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error)

func ListFileNames(s3api s3iface.S3API, bucket string, prefix string) (names []string, err error) {
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(10),
		Prefix:  aws.String(prefix),
	}
	output, err := s3api.ListObjects(input)
	if err != nil {
		return nil, err
	}
	names = make([]string, len(output.Contents))
	for idx, content := range output.Contents {
		names[idx] = *content.Key
	}
	return names, nil
}

func ListFileNamesExample() {
	sess := session.Must(session.NewSession())
	s3api := s3.New(sess)
	names, _ := ListFileNames(s3api, "examplebucket", "/a/b")
	fmt.Println(names)
}
