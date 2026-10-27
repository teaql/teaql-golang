package securedocumentreference

import (
	"strconv"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

// SecureDocumentAssembly supplies only trusted, current host state. It must
// never derive a snapshot or principal from submitted document JSON.
type SecureDocumentAssembly interface {
	Principal(*runtime.UserContext) (TrustedReferencePrincipal, error)
	Snapshot(*runtime.UserContext, runtime.AuthorizedResponseGraph) (PreparedBoundDocument, error)
}

type PreparedBoundDocument struct {
	Document BoundDocumentSnapshot
	Lifetime time.Duration
}

type SecureDocumentReferenceModule struct{ Protocol *ContextBoundReferenceRuntime }

type secureCapability struct{ contributor *secureContributor }

func (*secureCapability) Capability() string { return "secure_document.v1" }
func (c *secureCapability) ValidateReady(context *runtime.UserContext) error {
	if context == nil || c.contributor == nil || context.GetResource("secureDocumentContributor") != c.contributor || c.contributor.assembly == nil || c.contributor.module == nil || c.contributor.module.Protocol == nil {
		return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	return nil
}

type secureContributor struct {
	module     *SecureDocumentReferenceModule
	context    *runtime.UserContext
	assembly   SecureDocumentAssembly
	capability *secureCapability
}

func (m *SecureDocumentReferenceModule) Install(context *runtime.UserContext, assembly SecureDocumentAssembly) error {
	if m == nil || m.Protocol == nil || context == nil || assembly == nil || context.GetResource("secureDocumentContributor") != nil || context.GetResource("teaql.runtime.capability:secure_document.v1") != nil {
		return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	contributor := &secureContributor{module: m, context: context, assembly: assembly}
	capability := &secureCapability{contributor: contributor}
	contributor.capability = capability
	if err := context.InstallRuntimeCapability(capability); err != nil {
		return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	context.InsertResource("secureDocumentContributor", contributor)
	return nil
}

// Ready is an explicit startup check; Install only wires the optional module.
func (m *SecureDocumentReferenceModule) Ready(context *runtime.UserContext) error {
	if m == nil || m.Protocol == nil || context == nil {
		return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	capability, ok := context.GetResource("teaql.runtime.capability:secure_document.v1").(*secureCapability)
	if !ok || capability.contributor.module != m {
		return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	return capability.ValidateReady(context)
}

func projectionError() error {
	return &runtime.ViewPreparationError{Code: "DOCUMENT_PROJECTION_VIOLATION"}
}
func wrapDocumentError(err error) error {
	if reference, ok := err.(*ReferenceError); ok {
		return &runtime.ViewPreparationError{Code: reference.Code}
	}
	return &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
}

func (c *secureContributor) Prepare(context *runtime.UserContext, graph runtime.AuthorizedResponseGraph, options runtime.ViewOptions) (runtime.PreparedSecureDocument, error) {
	var empty runtime.PreparedSecureDocument
	if context != c.context || options.Mode != "governed" || c.capability.ValidateReady(context) != nil {
		return empty, &runtime.ViewPreparationError{Code: "DOCUMENT_CONFIGURATION"}
	}
	principal, err := c.assembly.Principal(context)
	if err != nil {
		return empty, wrapDocumentError(err)
	}
	prepared, err := c.assembly.Snapshot(context, graph)
	if err != nil {
		return empty, wrapDocumentError(err)
	}
	for _, slot := range graph.BlobSlots {
		id, err := strconv.ParseUint(slot.EntityID, 10, 64)
		if err != nil || id == 0 || strconv.FormatUint(id, 10) != slot.EntityID {
			return empty, projectionError()
		}
		rowKey := ""
		for _, entity := range prepared.Document.Snapshot.Entities {
			if entity.Identity.EntityType == slot.Entity && entity.Identity.ID == id {
				rowKey = entity.RowKey
				break
			}
		}
		if rowKey == "" {
			return empty, projectionError()
		}
		matched := false
		for _, binding := range prepared.Document.Bindings {
			if binding.RowKey == rowKey && binding.CanonicalField == slot.Field && binding.EntityType == slot.Entity && binding.StableKey != nil && *binding.StableKey == slot.Reference.Key() && binding.BindingVersion == slot.BindingVersion {
				matched = true
				break
			}
		}
		if !matched {
			return empty, projectionError()
		}
	}
	token, err := c.module.Protocol.IssueBoundDocumentSnapshot(principal, prepared.Document, prepared.Lifetime)
	if err != nil {
		return empty, wrapDocumentError(err)
	}
	verified, err := c.module.Protocol.ConsumeBoundDocumentSnapshot(principal, token, prepared.Document.Snapshot.Scope.AggregateType, prepared.Document.Snapshot.Scope.Purpose)
	if err != nil {
		return empty, wrapDocumentError(err)
	}
	return runtime.PreparedSecureDocument{Graph: graph, Envelope: map[string]any{"businessId": prepared.Document.Snapshot.BusinessID, "documentToken": token}, ExpiresAt: verified.ExpiresAt}, nil
}
