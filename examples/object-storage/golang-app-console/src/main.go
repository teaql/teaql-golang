package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teaql/teaql-golang/core"
	storageS3 "github.com/teaql/teaql-golang/objectstorage/s3"
	"github.com/teaql/teaql-golang/runtime"
	lib "order-attachment-service-core-workspace/lib"
	"order-attachment-service-core-workspace/lib/order_attachment"
)

const receipt = "TeaQL order receipt fixture\nItem: tea\nTotal: 12.50\n"
const mediaType = "text/plain"
const filename = "receipt.txt"

type orderAudit struct{ events []*runtime.SafeAuditEvent }

func (a *orderAudit) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if event.Entity == "customer_order" || event.Entity == "order_attachment" {
		a.events = append(a.events, event)
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func checked[T any](value T, err error) T {
	must(err)
	return value
}

func require(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func expect(code core.StorageErrorCode, err error) {
	var typed *core.StorageError
	require(errors.As(err, &typed) && typed.Code == code, "unexpected storage error")
}

func verify(database, download string, round int) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	defer httpClient.CloseIdleConnections()
	client := sdk.NewFromConfig(aws.Config{
		Region: "us-east-1", HTTPClient: httpClient,
		Credentials: credentials.NewStaticCredentialsProvider(os.Getenv("TEAQL_S3_ACCESS_KEY"), os.Getenv("TEAQL_S3_SECRET_KEY"), ""),
		Retryer:     func() aws.Retryer { return aws.NopRetryer{} },
	}, func(options *sdk.Options) {
		options.BaseEndpoint = aws.String(os.Getenv("TEAQL_S3_ENDPOINT"))
		options.UsePathStyle = true
	})
	provider := checked(storageS3.New(client, os.Getenv("TEAQL_S3_BUCKET")))
	must(os.Setenv("ORDER_ATTACHMENT_SERVICE_CORE_DATABASE_URL", database))
	context := checked(lib.ServiceRuntimeFromEnv())
	defer func() { must(context.GetResource("db").(*sql.DB).Close()) }()
	trace := fmt.Sprintf("order-attachment-%d", round)
	var events []runtime.ObjectStorageEvent
	context.WithTraceId(trace).WithObjectStorage(provider).WithObjectStorageObserver(
		runtime.ObjectStorageObserverFunc(func(event runtime.ObjectStorageEvent) {
			events = append(events, event)
		}))
	audit := &orderAudit{}
	context.WithAppAuditEventSink(audit)
	must(lib.EnsureSchema(context))
	platform := checked(lib.Q.CommercePlatforms().WithIdIs(1).Limit(1).
		Comment("load the commerce platform").Purpose("associate the new order to its root").ExecuteForOne(context))
	require(platform != nil, "generated platform missing")

	orderNumber := fmt.Sprintf("ATTACH-ORDER-%d", round)
	order := checked(lib.Q.CustomerOrders().Comment("prepare receipt order").
		Purpose("file a receipt against this order").NewEntity(context))
	order.UpdateOrderNumber(orderNumber).UpdateCommercePlatformId(platform.Id())
	order = checked(order.AuditAs("place the order receiving the receipt").Save(context))

	storage := checked(context.ObjectStorage("file and retrieve an order receipt"))
	random := make([]byte, 16)
	_, err := rand.Read(random)
	must(err)
	key := checked(core.NewObjectKey("private-order-attachments/" + hex.EncodeToString(random)))
	uploaded := false
	defer func() {
		if uploaded {
			must(storage.Delete(key))
		}
	}()
	require(!checked(storage.Exists(key)), "new attachment key already exists")
	body := []byte(receipt)
	length := int64(len(body))
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	put := checked(core.NewObjectPutOptions(length, core.ObjectCreateOnly))
	put = checked(put.WithContentType(mediaType))
	put = checked(put.WithMetadata(map[string]string{"receipt-sha256": digest}))
	checked(storage.PutBytes(key, body, put))
	uploaded = true

	attachment := checked(lib.Q.OrderAttachments().Comment("prepare receipt attachment").
		Purpose("save the uploaded receipt reference on its order").NewEntity(context))
	attachment.UpdateCustomerOrderId(order.Id()).UpdateFileName(filename).UpdateContentType(mediaType).
		UpdateContentLength(length).UpdateContentSha256(digest).UpdateObjectKey(key.Value())
	attachment = checked(attachment.AuditAs("associate the stored receipt with its order").Save(context))
	attachments := func() []*order_attachment.OrderAttachment {
		return checked(lib.Q.OrderAttachments().WithCustomerOrderIs(order.Id()).Limit(2).
			Comment("load order attachments").Purpose("retrieve this order's stored receipt").ExecuteForList(context)).Data
	}
	rows := attachments()
	require(len(rows) == 1 && rows[0].Id() == attachment.Id(), "persisted order association mismatch")
	expression := order_attachment.NewOrderAttachmentExpression(rows[0])
	parent, parentLoaded := expression.CustomerOrderId().Eval()
	file, fileLoaded := expression.FileName().Eval()
	contentType, typeLoaded := expression.ContentType().Eval()
	contentLength, lengthLoaded := expression.ContentLength().Eval()
	contentHash, hashLoaded := expression.ContentSha256().Eval()
	objectKey, keyLoaded := expression.ObjectKey().Eval()
	require(parentLoaded && fileLoaded && typeLoaded && lengthLoaded && hashLoaded && keyLoaded, "attachment projection incomplete")
	require(parent == order.Id() && file == filename && contentType == mediaType &&
		contentLength == length && contentHash == digest && objectKey == key.Value(), "stored attachment metadata mismatch")
	persistedKey := checked(core.NewObjectKey(objectKey))
	metadata := checked(storage.Head(persistedKey))
	require(metadata != nil && metadata.Size() == contentLength && metadata.ContentType() == contentType &&
		metadata.CustomMetadata()["receipt-sha256"] == contentHash, "object provider metadata mismatch")

	replaceAttempt := []byte("replacement")
	_, err = storage.PutBytes(persistedKey, replaceAttempt, checked(core.NewObjectPutOptions(int64(len(replaceAttempt)), core.ObjectCreateOnly)))
	expect(core.StorageConflict, err)
	_, err = storage.GetBytes(persistedKey, contentLength-1)
	expect(core.StorageTooLarge, err)
	returned := checked(storage.GetBytes(persistedKey, contentLength))
	require(bytes.Equal(returned, body) && fmt.Sprintf("%x", sha256.Sum256(returned)) == contentHash, "retrieved receipt changed")
	must(os.WriteFile(download, returned, 0o600))

	must(storage.Delete(persistedKey))
	uploaded = false
	_, err = storage.Get(persistedKey)
	expect(core.StorageNotFound, err)
	require(!checked(storage.Exists(persistedKey)), "missing attachment reported present")
	rows = attachments()
	require(len(rows) == 1 && rows[0].Id() == attachment.Id(), "missing object changed the order association")
	require(len(events) > 0, "storage observations absent")
	for _, event := range events {
		require(event.TraceID == trace, "storage observation escaped the attachment Context")
	}
	require(len(audit.events) == 2 && audit.events[0].Entity == "customer_order" && audit.events[1].Entity == "order_attachment",
		"committed order and attachment audits missing")
	require(!strings.Contains(string(checked(json.Marshal(events))), key.Value()) &&
		!strings.Contains(string(checked(json.Marshal(audit.events))), key.Value()), "private object key reached safe events")
	result := map[string]any{
		"round": round, "order": order.Id(), "attachment": attachment.Id(), "orderNumber": orderNumber,
		"filename": filename, "contentType": mediaType, "length": length, "sha256": digest,
		"conflict": true, "sizeRefused": true, "missing": true, "businessAudits": len(audit.events),
	}
	fmt.Println("ORDER_ATTACHMENT " + string(checked(json.Marshal(result))))
}

func main() {
	require(len(os.Args) == 4, "provide database, downloaded receipt path and round")
	verify(os.Args[1], os.Args[2], checked(strconv.Atoi(os.Args[3])))
}
