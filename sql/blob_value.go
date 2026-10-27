package sql

import (
	"errors"

	"github.com/teaql/teaql-golang/core"
)

var errBlobCapabilityMissing = errors.New("BLOB_CAPABILITY_MISSING")

// A modeled Blob column accepts only a trusted typed reference or an allowed null.
// A raw Key string, including one received through a generic request, is never a bind credential.
func validateBlobMutationValue(entity *core.EntityDescriptor, property *core.PropertyDescriptor, value core.Value) error {
	if property.DataType != core.TypeBlobRef && property.BlobSpec == nil {
		return nil
	}
	spec := property.BlobSpec
	if property.DataType != core.TypeBlobRef || spec == nil || spec.Entity != entity.Name ||
		spec.Field != property.Name || len(spec.AllowedExtensions()) == 0 || spec.Nullable != property.Nullable {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	if value.V == nil || value.V == core.TypeBlobRef {
		if property.Nullable {
			return nil
		}
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	reference, ok := value.V.(core.BlobRef)
	if !ok {
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	if _, err := reference.Value(); err != nil {
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	return nil
}
