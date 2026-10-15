package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teaql/teaql-golang/core"
	storageS3 "github.com/teaql/teaql-golang/objectstorage/s3"
	"github.com/teaql/teaql-golang/runtime"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func require(ok bool) {
	if !ok {
		panic("object storage contract assertion failed")
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}
func expect(code core.StorageErrorCode, err error) {
	var typed *core.StorageError
	require(errors.As(err, &typed) && typed.Code == code)
}
func key(value string) core.ObjectKey { k, err := core.NewObjectKey(value); check(err); return k }
func options(size int64, mode core.ObjectWriteMode) core.ObjectPutOptions {
	o, err := core.NewObjectPutOptions(size, mode)
	check(err)
	return o
}
func main() {
	bucket := os.Getenv("TEAQL_S3_BUCKET")
	httpClient := &http.Client{Timeout: 30 * time.Second}
	defer httpClient.CloseIdleConnections()
	client := sdk.NewFromConfig(aws.Config{Region: "us-east-1", HTTPClient: httpClient, Credentials: credentials.NewStaticCredentialsProvider(os.Getenv("TEAQL_S3_ACCESS_KEY"), os.Getenv("TEAQL_S3_SECRET_KEY"), ""), Retryer: func() aws.Retryer { return aws.NopRetryer{} }}, func(o *sdk.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("TEAQL_S3_ENDPOINT"))
		o.UsePathStyle = true
	})
	_, err := client.HeadBucket(context.Background(), &sdk.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		var response interface{ HTTPStatusCode() int }
		if !errors.As(err, &response) || response.HTTPStatusCode() != 404 {
			panic("test bucket setup failed")
		}
		_, err = client.CreateBucket(context.Background(), &sdk.CreateBucketInput{Bucket: aws.String(bucket)})
		check(err)
	}
	provider, err := storageS3.New(client, bucket)
	check(err)
	var mu sync.Mutex
	events := []runtime.ObjectStorageEvent{}
	observer := runtime.ObjectStorageObserverFunc(func(e runtime.ObjectStorageEvent) { mu.Lock(); defer mu.Unlock(); events = append(events, e) })
	user := runtime.NewUserContext().WithTraceId("storage-example").WithObjectStorage(provider).WithObjectStorageObserver(observer)
	storage, err := user.ObjectStorage("verify binary object storage")
	check(err)
	random := make([]byte, 16)
	_, err = rand.Read(random)
	check(err)
	prefix := "go-example/" + hex.EncodeToString(random) + "/"
	binary, empty, race := key(prefix+"收据 +% binary"), key(prefix+"empty"), key(prefix+"race")
	defer func() {
		for _, k := range []core.ObjectKey{binary, empty, race} {
			_ = storage.Delete(k)
		}
	}()
	exists, err := storage.Exists(binary)
	check(err)
	require(!exists)
	_, err = storage.Get(binary)
	expect(core.StorageNotFound, err)
	data := make([]byte, 256*1024)
	for n := range data {
		data[n] = byte(n % 251)
	}
	config, err := options(int64(len(data)), core.ObjectCreateOnly).WithMetadata(map[string]string{"example": "binary"})
	check(err)
	_, err = storage.PutBytes(binary, data, config)
	check(err)
	meta, err := storage.Head(binary)
	check(err)
	require(meta.Size() == int64(len(data)) && meta.CustomMetadata()["example"] == "binary" && meta.ETag() != "")
	returned, err := storage.GetBytes(binary, int64(len(data)))
	check(err)
	require(bytes.Equal(returned, data))
	_, err = storage.PutBytes(binary, nil, options(0, core.ObjectCreateOnly))
	expect(core.StorageConflict, err)
	_, err = storage.GetBytes(binary, 1)
	expect(core.StorageTooLarge, err)
	urlOptions, err := core.NewObjectReadURLOptions(5*time.Minute, "attachment", "receipt.bin")
	check(err)
	link, err := storage.GetReadURL(binary, urlOptions)
	check(err)
	response, err := httpClient.Get(link.URL())
	check(err)
	returned, err = io.ReadAll(response.Body)
	check(err)
	_ = response.Body.Close()
	require(response.StatusCode == 200 && bytes.Equal(returned, data) && response.Header.Get("Content-Disposition") == "attachment; filename=\"receipt.bin\"")
	urlOptions, err = core.NewObjectReadURLOptions(time.Second, "", "")
	check(err)
	expired, err := storage.GetReadURL(binary, urlOptions)
	check(err)
	time.Sleep(2200 * time.Millisecond)
	response, err = httpClient.Get(expired.URL())
	check(err)
	_ = response.Body.Close()
	require(response.StatusCode == 403)
	_, err = storage.PutBytes(empty, nil, options(0, core.ObjectCreateOnly))
	check(err)
	returned, err = storage.GetBytes(empty, 0)
	check(err)
	require(len(returned) == 0)
	_, err = storage.PutBytes(binary, []byte("ab"), options(1, core.ObjectReplace))
	expect(core.StorageInvalidInput, err)
	returned, err = storage.GetBytes(binary, int64(len(data)))
	check(err)
	require(bytes.Equal(returned, data))
	_, err = storage.PutBytes(binary, []byte("new"), options(3, core.ObjectReplace))
	check(err)
	returned, err = storage.GetBytes(binary, 3)
	check(err)
	require(string(returned) == "new")
	other := runtime.NewUserContext().WithTraceId("storage-other").WithObjectStorage(provider).WithObjectStorageObserver(observer)
	otherStorage, err := other.ObjectStorage("verify atomic create")
	check(err)
	outcomes := make(chan error, 2)
	for _, api := range []*runtime.ObjectStorage{storage, otherStorage} {
		go func(api *runtime.ObjectStorage) {
			_, err := api.PutBytes(race, []byte("race"), options(4, core.ObjectCreateOnly))
			outcomes <- err
		}(api)
	}
	wins := 0
	for n := 0; n < 2; n++ {
		err := <-outcomes
		if err == nil {
			wins++
		} else {
			expect(core.StorageConflict, err)
		}
	}
	require(wins == 1)
	missingProvider, err := storageS3.New(client, "missing-"+hex.EncodeToString(random))
	check(err)
	missing, err := runtime.NewUserContext().WithObjectStorage(missingProvider).ObjectStorage("verify missing bucket")
	check(err)
	_, err = missing.Exists(binary)
	expect(core.StorageUnavailable, err)
	for _, k := range []core.ObjectKey{binary, empty, race} {
		check(storage.Delete(k))
		check(storage.Delete(k))
		exists, err = storage.Exists(k)
		check(err)
		require(!exists)
	}
	seen := map[string]bool{}
	for _, e := range events {
		seen[e.TraceID] = true
	}
	require(seen["storage-example"] && seen["storage-other"])
	fmt.Println("PASS Go S3: binary, empty, metadata, URL download/expiry, atomic create, replace, bounded reads, missing bucket, repeated delete")
}
