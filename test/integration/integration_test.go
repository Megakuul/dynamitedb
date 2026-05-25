package integration

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/megakuul/dynamitedb"
)

type Test struct {
	PartId       dynamitedb.KeyField                     `pk:"part" json:"-"`
	SortId       dynamitedb.KeyField                     `sk:"sort" json:"-"`
	Nested       *NestedTest                             `json:"nested,omitempty"`
	TestString   dynamitedb.DataField[string]            `json:"test_string,omitempty"`
	TestInt      dynamitedb.DataField[int]               `json:"test_int,omitempty"`
	TestFloat    dynamitedb.DataField[float64]           `json:"test_float,omitempty"`
	TestSlice    dynamitedb.DataField[[]string]          `json:"test_slice,omitempty"`
	TestMap      dynamitedb.DataField[map[string]string] `json:"test_map,omitempty"`
	TestBool     dynamitedb.DataField[bool]              `json:"test_bool,omitempty"`
	TestTime     dynamitedb.DataField[time.Time]         `json:"test_time,omitempty"`
	TestDuration dynamitedb.DataField[time.Duration]     `json:"test_duration,omitempty"`

	TestUnmodified dynamitedb.DataField[string]            `json:"test_unmodified,omitempty"`
	TestNil        dynamitedb.DataField[string]            `json:"test_nil,omitempty"`
	TestNilMap     dynamitedb.DataField[map[string]string] `json:"test_nil_map,omitempty"`
}

type NestedTest struct {
	TestString dynamitedb.DataField[string] `json:"test_string,omitempty"`
	Nested     NestedNestedTest             `json:"nested"`
}

type NestedNestedTest struct {
	TestString dynamitedb.DataField[string] `json:"test_string,omitempty"`
}

func TestOperations(t *testing.T) {
	// prepare
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	server := httptest.NewServer(faker.Server())
	defer server.Close()

	cfg, err := config.LoadDefaultConfig(
		t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("ACCESS_KEY", "SECRET_KEY", "")),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(server.URL)
		o.UsePathStyle = true
	})
	_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String("test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	bucket := dynamitedb.NewFromClient(client, "test")

	t.Run("basic workflow", func(t *testing.T) {
		checkWorkflow(t, bucket)
	})

	t.Run("insert operations", func(t *testing.T) {
		checkInserts(t, bucket)
	})

	t.Run("query operations", func(t *testing.T) {
		checkQueries(t, bucket)
	})
}
