package blobreference

import (
	"io"
	"reflect"
	"sort"
)

// ContentVerifierRoute groups extensions served by one optional parser.
// The host selects these routes explicitly; construction performs no I/O.
type ContentVerifierRoute struct {
	Extensions []string
	Verifier   ContentVerifier
}

// CompositeContentVerifier connects independent parsers to the single trusted
// upload verifier SPI. Unregistered extensions are rejected before source I/O.
type CompositeContentVerifier struct {
	routes map[string]ContentVerifier
	ready  []ContentVerifier
}

func validVerifierExtension(extension string) bool {
	if len(extension) < 1 || len(extension) > 16 {
		return false
	}
	for index := 0; index < len(extension); index++ {
		value := extension[index]
		if value < 'a' || value > 'z' {
			if value < '0' || value > '9' {
				return false
			}
		}
	}
	return true
}

func NewCompositeContentVerifier(routes ...ContentVerifierRoute) (*CompositeContentVerifier, error) {
	if len(routes) < 1 || len(routes) > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	result := &CompositeContentVerifier{routes: make(map[string]ContentVerifier)}
	for _, route := range routes {
		if route.Verifier == nil || len(route.Extensions) < 1 || len(route.Extensions) > 32 {
			return nil, uploadError("BLOB_POLICY_INVALID")
		}
		value := reflect.ValueOf(route.Verifier)
		switch value.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if value.IsNil() {
				return nil, uploadError("BLOB_POLICY_INVALID")
			}
		}
		for _, extension := range route.Extensions {
			if !validVerifierExtension(extension) || len(result.routes) >= 32 || result.routes[extension] != nil {
				return nil, uploadError("BLOB_POLICY_INVALID")
			}
			result.routes[extension] = route.Verifier
		}
		result.ready = append(result.ready, route.Verifier)
	}
	return result, nil
}

func (v *CompositeContentVerifier) SupportedExtensions() []string {
	if v == nil {
		return nil
	}
	extensions := make([]string, 0, len(v.routes))
	for extension := range v.routes {
		extensions = append(extensions, extension)
	}
	sort.Strings(extensions)
	return extensions
}

func (v *CompositeContentVerifier) ValidateReady() error {
	if v == nil {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	for _, verifier := range v.ready {
		if ready, ok := verifier.(interface{ ValidateReady() error }); ok {
			if err := ready.ValidateReady(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *CompositeContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	if v == nil || v.routes[extension] == nil {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return v.routes[extension].Verify(source, extension, maximumBytes)
}
