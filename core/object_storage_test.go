package core

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestObjectStorageValidationAndPrivateFormatting(t *testing.T) {
	for _, key := range []string{"", strings.Repeat("x", 1025), "bad\nkey", string([]byte{0xff})} {
		if _, err := NewObjectKey(key); err == nil {
			t.Fatal("invalid key accepted")
		}
	}
	for _, size := range []int64{-1, MaxObjectPutSize + 1} {
		if _, err := NewObjectPutOptions(size, ObjectCreateOnly); err == nil {
			t.Fatal("invalid size accepted")
		}
	}
	options, _ := NewObjectPutOptions(0, ObjectCreateOnly)
	if _, err := options.WithContentType("bad\r\nheader"); err == nil {
		t.Fatal("unsafe header accepted")
	}
	if _, err := options.WithMetadata(map[string]string{"Bad Key": "value"}); err == nil {
		t.Fatal("unsafe metadata accepted")
	}
	original := map[string]string{"tag": "PRIVATE-METADATA-CANARY"}
	options, _ = options.WithMetadata(original)
	original["tag"] = "changed"
	copy := options.Metadata()
	copy["tag"] = "changed"
	if options.Metadata()["tag"] != "PRIVATE-METADATA-CANARY" {
		t.Fatal("metadata was mutable")
	}
	for _, lifetime := range []time.Duration{0, -time.Second, time.Nanosecond, 8 * 24 * time.Hour} {
		if _, err := NewObjectReadURLOptions(lifetime, "", ""); err == nil {
			t.Fatal("invalid lifetime accepted")
		}
	}
	if _, err := NewObjectReadURLOptions(time.Second, "attachment", "bad\nname"); err == nil {
		t.Fatal("unsafe filename accepted")
	}
	key, _ := NewObjectKey("PRIVATE-KEY-CANARY")
	url := NewObjectReadURL("PRIVATE-URL-CANARY", time.Now())
	metadata, _ := NewObjectMetadata(key, 0, "PRIVATE-TYPE-CANARY", "PRIVATE-ETAG-CANARY", nil, original)
	for _, value := range []interface{}{key, url, metadata, options} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			if strings.Contains(fmt.Sprintf(format, value), "CANARY") {
				t.Fatal("private diagnostic value exposed")
			}
		}
	}
}
