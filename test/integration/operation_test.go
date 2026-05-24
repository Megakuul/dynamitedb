package integration

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/megakuul/dynamitedb"
)

type Test struct {
	PartId     dynamitedb.KeyField                     `pk:"part" json:"-"`
	SortId     dynamitedb.KeyField                     `sk:"sort" json:"-"`
	Nested     *NestedTest                             `json:"nested"`
	TestString dynamitedb.DataField[string]            `json:"test_string"`
	TestInt    dynamitedb.DataField[int]               `json:"test_int"`
	TestFloat  dynamitedb.DataField[float64]           `json:"test_float"`
	TestSlice  dynamitedb.DataField[[]string]          `json:"test_slice"`
	TestMap    dynamitedb.DataField[map[string]string] `json:"test_map"`
	TestBool   dynamitedb.DataField[bool]              `json:"test_bool"`

	TestUnmodified dynamitedb.DataField[string]            `json:"test_unmodified"`
	TestNil        dynamitedb.DataField[string]            `json:"test_nil"`
	TestNilMap     dynamitedb.DataField[map[string]string] `json:"test_nil_map"`
}

type NestedTest struct {
	TestString dynamitedb.DataField[string] `json:"test_string"`
	Nested     NestedNestedTest             `json:"nested"`
}

type NestedNestedTest struct {
	TestString dynamitedb.DataField[string] `json:"test_string"`
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

	err = dynamitedb.Create(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("1"),
		SortId:     dynamitedb.Key("50"),
		TestString: dynamitedb.Set("Bombaclad"),
	})
	if err != nil {
		t.Fatal(err)
	}

	output, err := client.ListObjectsV2(t.Context(), &s3.ListObjectsV2Input{Bucket: aws.String("test")})
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range output.Contents {
		println(*content.Key)
	}

	res, err := dynamitedb.Get(t.Context(), bucket, &Test{
		PartId: dynamitedb.Key("1"),
		SortId: dynamitedb.Key("50"),
	})
	if err != nil {
		t.Fatal(err)
	}

	println(res.TestString.Value())

	// TODO
}
