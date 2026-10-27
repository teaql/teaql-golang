// Package s3 implements the provider SPI using the official AWS Go SDK v2.
package s3

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Provider struct {
	client *sdk.Client
	signer *sdk.PresignClient
	bucket string
}

func New(client *sdk.Client, bucket string) (*Provider, error) {
	if client == nil || strings.TrimSpace(bucket) == "" {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	return &Provider{client, sdk.NewPresignClient(client), bucket}, nil
}
func (p *Provider) String() string   { return "S3ObjectStorage[configuration=masked]" }
func (p *Provider) GoString() string { return p.String() }
func translate(err error) error {
	if err == nil {
		return nil
	}
	code := core.StorageUnavailable
	var typed *core.StorageError
	if errors.As(err, &typed) {
		return &core.StorageError{Code: typed.Code}
	}
	var timeout net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &timeout) && timeout.Timeout() {
		return &core.StorageError{Code: core.StorageTimeout}
	}
	var api smithy.APIError
	if errors.As(err, &api) && api.ErrorCode() == "NoSuchBucket" {
		return &core.StorageError{Code: core.StorageUnavailable}
	}
	var response *smithyhttp.ResponseError
	if errors.As(err, &response) {
		switch response.HTTPStatusCode() {
		case 401, 403:
			code = core.StorageAccessDenied
		case 404:
			code = core.StorageNotFound
		case 409, 412:
			code = core.StorageConflict
		case 408, 504:
			code = core.StorageTimeout
		case 413:
			code = core.StorageTooLarge
		case 501:
			code = core.StorageUnsupported
		case 400, 422:
			code = core.StorageInvalidInput
		}
	}
	return &core.StorageError{Code: code}
}
func (p *Provider) Put(i runtime.StorageInvocation, key core.ObjectKey, content io.Reader, options core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	staging, err := os.CreateTemp("", "teaql-object-*.upload")
	if err != nil {
		return nil, &core.StorageError{Code: core.StorageIO}
	}
	defer os.Remove(staging.Name())
	defer staging.Close()
	// Read one byte beyond the declared size to detect excess before any request.
	copied, err := io.CopyBuffer(staging, io.LimitReader(content, options.Size()+1), make([]byte, 65536))
	if err != nil {
		return nil, &core.StorageError{Code: core.StorageIO}
	}
	if copied != options.Size() {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	if _, err = staging.Seek(0, io.SeekStart); err != nil {
		return nil, &core.StorageError{Code: core.StorageIO}
	}
	input := &sdk.PutObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key.Value()), Body: staging, ContentLength: aws.Int64(options.Size()), ContentType: aws.String(options.ContentType()), Metadata: options.Metadata()}
	if options.Mode() == core.ObjectCreateOnly {
		input.IfNoneMatch = aws.String("*")
	}
	result, err := p.client.PutObject(i.Context(), input)
	if err != nil {
		return nil, translate(err)
	}
	if result.ETag == nil || strings.TrimSpace(*result.ETag) == "" {
		return nil, &core.StorageError{Code: core.StorageUnavailable}
	}
	return core.NewObjectMetadata(key, options.Size(), options.ContentType(), *result.ETag, nil, options.Metadata())
}
func metadata(key core.ObjectKey, size *int64, contentType, etag *string, modified *time.Time, values map[string]string) (*core.ObjectMetadata, error) {
	if size == nil || etag == nil || strings.TrimSpace(*etag) == "" {
		return nil, &core.StorageError{Code: core.StorageUnavailable}
	}
	return core.NewObjectMetadata(key, *size, aws.ToString(contentType), *etag, modified, values)
}
func (p *Provider) Get(i runtime.StorageInvocation, key core.ObjectKey) (*runtime.ObjectContent, error) {
	result, err := p.client.GetObject(i.Context(), &sdk.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key.Value())})
	if err != nil {
		return nil, translate(err)
	}
	meta, err := metadata(key, result.ContentLength, result.ContentType, result.ETag, result.LastModified, result.Metadata)
	if err != nil {
		_ = result.Body.Close()
		return nil, err
	}
	return runtime.NewObjectContent(meta, result.Body), nil
}
func (p *Provider) Head(i runtime.StorageInvocation, key core.ObjectKey) (*core.ObjectMetadata, error) {
	result, err := p.client.HeadObject(i.Context(), &sdk.HeadObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key.Value())})
	if err == nil {
		return metadata(key, result.ContentLength, result.ContentType, result.ETag, result.LastModified, result.Metadata)
	}
	mapped := translate(err)
	var typed *core.StorageError
	if errors.As(mapped, &typed) && typed.Code == core.StorageNotFound {
		_, bucketErr := p.client.HeadBucket(i.Context(), &sdk.HeadBucketInput{Bucket: aws.String(p.bucket)})
		if bucketErr == nil {
			return nil, nil
		}
		mapped = translate(bucketErr)
		if errors.As(mapped, &typed) && typed.Code == core.StorageNotFound {
			return nil, &core.StorageError{Code: core.StorageUnavailable}
		}
	}
	return nil, mapped
}
func (p *Provider) Delete(i runtime.StorageInvocation, key core.ObjectKey) error {
	_, err := p.client.DeleteObject(i.Context(), &sdk.DeleteObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key.Value())})
	return translate(err)
}
func (p *Provider) DeleteIfVersion(i runtime.StorageInvocation, key core.ObjectKey, contentVersion string) (runtime.ObjectConditionalDeleteResult, error) {
	if !key.Valid() || strings.TrimSpace(contentVersion) != contentVersion || contentVersion == "" || contentVersion == "*" || len(contentVersion) > 256 || strings.IndexFunc(contentVersion, unicode.IsControl) >= 0 {
		return "", &core.StorageError{Code: core.StorageInvalidInput}
	}
	_, err := p.client.DeleteObject(i.Context(), &sdk.DeleteObjectInput{
		Bucket: aws.String(p.bucket), Key: aws.String(key.Value()), IfMatch: aws.String(contentVersion),
	})
	if err == nil {
		return runtime.ObjectDeleted, nil
	}
	var response *smithyhttp.ResponseError
	if errors.As(err, &response) && (response.HTTPStatusCode() == 404 || response.HTTPStatusCode() == 412) {
		current, headErr := p.Head(i, key)
		if headErr != nil {
			return "", headErr
		}
		if current == nil {
			return runtime.ObjectAlreadyMissing, nil
		}
		return runtime.ObjectVersionMismatch, nil
	}
	return "", translate(err)
}
func (p *Provider) GetReadURL(i runtime.StorageInvocation, key core.ObjectKey, options core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	input := &sdk.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key.Value())}
	if options.ContentDisposition() != "" {
		input.ResponseContentDisposition = aws.String(options.ContentDisposition())
	}
	result, err := p.signer.PresignGetObject(i.Context(), input, func(o *sdk.PresignOptions) { o.Expires = options.Lifetime() })
	if err != nil {
		return core.ObjectReadURL{}, translate(err)
	}
	for header := range result.SignedHeader {
		if !strings.EqualFold(header, "host") {
			return core.ObjectReadURL{}, &core.StorageError{Code: core.StorageUnsupported}
		}
	}
	parsed, parseErr := url.Parse(result.URL)
	if parseErr != nil {
		return core.ObjectReadURL{}, &core.StorageError{Code: core.StorageUnavailable}
	}
	started, dateErr := time.Parse("20060102T150405Z", parsed.Query().Get("X-Amz-Date"))
	seconds, expiryErr := strconv.ParseInt(parsed.Query().Get("X-Amz-Expires"), 10, 64)
	if dateErr != nil || expiryErr != nil || seconds <= 0 || seconds > int64(options.Lifetime()/time.Second) {
		return core.ObjectReadURL{}, &core.StorageError{Code: core.StorageUnavailable}
	}
	return core.NewObjectReadURL(result.URL, started.Add(time.Duration(seconds)*time.Second)), nil
}
