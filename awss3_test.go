package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
)

type MockS3API struct {
	s3iface.S3API
	mock.Mock
}

func (m *MockS3API) ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.ListObjectsOutput), args.Error(1)
}

func TestListFileNames(t *testing.T) {
	tests := []struct {
		output   *s3.ListObjectsOutput
		err      error
		hasError bool
		count    int
		first    string
	}{
		{
			output: &s3.ListObjectsOutput{
				Contents: []*s3.Object{
					{Key: aws.String("/a/b/1.txt")},
					{Key: aws.String("/a/b/2.txt")},
				}},
			err:      nil,
			hasError: false,
			count:    2,
			first:    "/a/b/1.txt",
		},
		{
			output:   nil,
			err:      fmt.Errorf("bad network"),
			hasError: true,
		},
	}

	for row, test := range tests {
		s3api := new(MockS3API)
		s3api.On("ListObjects", mock.Anything).Return(test.output, test.err)

		names, err := ListFileNames(s3api, "anybucket", "anyprefix")
		if test.hasError {
			assert.Error(t, err, "row: %d", row)
			continue
		}
		assert.NoError(t, err, "row: %d", row)
		assert.Equal(t, test.count, len(names), "names count, row: %d", row)
		assert.Equal(t, test.first, names[0])
	}
}
