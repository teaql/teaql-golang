package blobreference

import (
	"errors"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

// BlobReferenceModule installs the optional SQLite catalog, verified upload
// service and authorized view projector into a host context without I/O.
// Schema migration remains an explicit deployment operation.
type BlobReferenceModule struct {
	Catalog *SQLiteCatalog
	Upload  *SQLiteUploadService
	Read    *SQLiteReadContributor
}

type blobCapability struct {
	module  *BlobReferenceModule
	context *runtime.UserContext
}

func (*blobCapability) Capability() string { return "blob.v1" }
func (c *blobCapability) ValidateReady(context *runtime.UserContext) error {
	if c == nil || c.module == nil || context == nil || context != c.context || c.module.Catalog == nil || c.module.Upload == nil || c.module.Read == nil ||
		c.module.Upload.catalog != c.module.Catalog || c.module.Read.Catalog != c.module.Catalog ||
		context.GetResource("blobResponseContributor") != c.module.Read || context.GetResource("blobUploadService") != c.module.Upload {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	if c.module.Read.Authorize == nil || c.module.Upload.authorization == nil || c.module.Upload.verifier == nil {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	if err := c.module.Catalog.ValidateSchema(context); err != nil {
		return err
	}
	if _, err := context.ObjectStorage("validate BlobReferenceModule readiness"); err != nil {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	return nil
}

func (m *BlobReferenceModule) Install(context *runtime.UserContext) error {
	if m == nil || context == nil || m.Catalog == nil || m.Upload == nil || m.Read == nil || m.Upload.catalog != m.Catalog || m.Read.Catalog != m.Catalog ||
		m.Catalog.transport == nil || m.Upload.authorization == nil || m.Upload.verifier == nil || m.Read.Authorize == nil ||
		context.GetResource("blobResponseContributor") != nil || context.GetResource("blobUploadService") != nil || context.GetResource("teaql.runtime.capability:blob.v1") != nil {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	capability := &blobCapability{module: m, context: context}
	if err := context.InstallRuntimeCapability(capability); err != nil {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	context.InsertResource("blobResponseContributor", m.Read)
	context.InsertResource("blobUploadService", m.Upload)
	return nil
}

func (m *BlobReferenceModule) Ready(context *runtime.UserContext) error {
	if m == nil || context == nil {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	capability, ok := context.GetResource("teaql.runtime.capability:blob.v1").(*blobCapability)
	if !ok || capability.module != m {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	return capability.ValidateReady(context)
}

// NewBlobReferenceModule checks assembly coherence before passive installation.
func NewBlobReferenceModule(catalog *SQLiteCatalog, upload *SQLiteUploadService,
	readAuthorization ReadAuthorization, maxURLAge time.Duration) (*BlobReferenceModule, error) {
	if catalog == nil || catalog.transport == nil || upload == nil || upload.catalog != catalog || upload.authorization == nil || upload.verifier == nil || readAuthorization == nil || maxURLAge < time.Second || maxURLAge > time.Hour || maxURLAge%time.Second != 0 {
		return nil, errors.New("BLOB_CAPABILITY_MISSING")
	}
	return &BlobReferenceModule{Catalog: catalog, Upload: upload, Read: &SQLiteReadContributor{Catalog: catalog, Authorize: readAuthorization, MaxURLAge: maxURLAge}}, nil
}
