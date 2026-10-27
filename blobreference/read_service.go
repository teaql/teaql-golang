package blobreference

import (
	"strings"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type ReadAuthorization interface {
	AuthorizeRead(*runtime.UserContext, runtime.AuthorizedBlobSlot) error
}

// SQLiteReadContributor reauthorizes current slots, then returns complete storage URLs.
type SQLiteReadContributor struct {
	Catalog   *SQLiteCatalog
	Authorize ReadAuthorization
	MaxURLAge time.Duration
}

func (c SQLiteReadContributor) ProjectBatch(actor *runtime.UserContext,
	graph runtime.AuthorizedResponseGraph, expiresAt time.Time) ([]runtime.ProjectedBlobURL, error) {
	if c.Catalog == nil || c.Authorize == nil || actor == nil {
		return nil, &runtime.ViewPreparationError{Code: "BLOB_CAPABILITY_MISSING"}
	}
	maximum := c.MaxURLAge
	if maximum == 0 {
		maximum = 10 * time.Minute
	}
	if maximum < time.Second || maximum > time.Hour || maximum%time.Second != 0 {
		return nil, &runtime.ViewPreparationError{Code: "BLOB_CAPABILITY_MISSING"}
	}
	urls := make([]runtime.ProjectedBlobURL, 0, len(graph.BlobSlots))
	for _, slot := range graph.BlobSlots {
		if err := c.Authorize.AuthorizeRead(actor, slot); err != nil {
			return nil, &runtime.ViewPreparationError{Code: "BLOB_READ_DENIED"}
		}
		record, err := c.Catalog.CurrentRead(actor, slot)
		if err != nil {
			return nil, &runtime.ViewPreparationError{Code: err.Error()}
		}
		lifetime := maximum
		if !expiresAt.IsZero() {
			limited := time.Duration(int64(expiresAt.Sub(time.Now()).Seconds())-1) * time.Second
			if limited < lifetime {
				lifetime = limited
			}
		}
		if lifetime < time.Second {
			return nil, &runtime.ViewPreparationError{Code: "BLOB_REFERENCE_EXPIRED"}
		}
		disposition := "attachment"
		if strings.HasPrefix(record.ContentType, "image/") || strings.HasPrefix(record.ContentType, "audio/") ||
			strings.HasPrefix(record.ContentType, "video/") || record.ContentType == "application/pdf" {
			disposition = "inline"
		}
		key, keyErr := core.NewObjectKey(record.ObjectKey)
		options, optionsErr := core.NewObjectReadURLOptions(lifetime, disposition, record.FileName)
		if keyErr != nil || optionsErr != nil {
			return nil, &runtime.ViewPreparationError{Code: "BLOB_SIGN_FAILED"}
		}
		store, storeErr := actor.ObjectStorage("prepare authorized Blob view")
		if storeErr != nil {
			return nil, &runtime.ViewPreparationError{Code: "BLOB_SIGN_FAILED"}
		}
		signed, signErr := store.GetReadURL(key, options)
		if signErr != nil {
			return nil, &runtime.ViewPreparationError{Code: "BLOB_SIGN_FAILED"}
		}
		urls = append(urls, runtime.ProjectedBlobURL{URL: signed.URL(), ExpiresAt: signed.ExpiresAt()})
	}
	return urls, nil
}
