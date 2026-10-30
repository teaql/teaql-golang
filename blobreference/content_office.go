package blobreference

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"io"
	"path"
	"strings"
	"unicode/utf8"
)

const (
	officePackageNamespace  = "http://schemas.openxmlformats.org/package/2006/"
	officeDocumentNamespace = "http://schemas.openxmlformats.org/officeDocument/2006/"
)

// OfficeContentVerifier is a passive, bounded DOCX/XLSX package and XML parser.
// It uses only Go's standard ZIP/XML libraries and never trusts filename MIME.
type OfficeContentVerifier struct {
	maximumInputBytes, maximumDecodedBytes int64
	maximumParts, maximumElements          int
	permits                                chan struct{}
}

func NewOfficeContentVerifier(maximumInputBytes, maximumDecodedBytes int64,
	maximumParts, maximumElements, concurrentVerifications int) (*OfficeContentVerifier, error) {
	if maximumInputBytes < 1 || maximumInputBytes > 100<<20 ||
		maximumDecodedBytes < 1 || maximumDecodedBytes > 100<<20 ||
		maximumParts < 1 || maximumParts > 10000 ||
		maximumElements < 1 || maximumElements > 2000000 ||
		concurrentVerifications < 1 || concurrentVerifications > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	return &OfficeContentVerifier{maximumInputBytes, maximumDecodedBytes,
		maximumParts, maximumElements, make(chan struct{}, concurrentVerifications)}, nil
}

func officePartName(name string) bool {
	if name == "" || len(name) > 1024 || !utf8.ValidString(name) ||
		strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	for _, character := range name {
		if character < 32 || character == 127 {
			return false
		}
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || component == "." || component == ".." {
			return false
		}
	}
	return true
}

func completeOfficeZIP(data []byte) bool {
	if len(data) < 22 || !bytes.HasPrefix(data, []byte{'P', 'K', 3, 4}) {
		return false
	}
	minimum := len(data) - 22 - 65535
	if minimum < 0 {
		minimum = 0
	}
	for index := len(data) - 22; index >= minimum; index-- {
		if !bytes.Equal(data[index:index+4], []byte{'P', 'K', 5, 6}) ||
			index+22+int(binary.LittleEndian.Uint16(data[index+20:])) != len(data) {
			continue
		}
		if binary.LittleEndian.Uint16(data[index+4:]) != 0 ||
			binary.LittleEndian.Uint16(data[index+6:]) != 0 ||
			binary.LittleEndian.Uint16(data[index+8:]) != binary.LittleEndian.Uint16(data[index+10:]) {
			return false
		}
		centralSize := uint64(binary.LittleEndian.Uint32(data[index+12:]))
		centralOffset := uint64(binary.LittleEndian.Uint32(data[index+16:]))
		return centralSize > 0 && centralOffset+centralSize == uint64(index)
	}
	return false
}

func officeAttribute(element xml.StartElement, name, namespace string) string {
	for _, attribute := range element.Attr {
		if attribute.Name.Local == name && attribute.Name.Space == namespace {
			return attribute.Value
		}
	}
	return ""
}

// scanOfficeXML parses the complete XML document, rejecting DTDs, extra roots,
// unbalanced tags and excessive depth/elements before any semantic promotion.
func scanOfficeXML(data []byte, count *int, maximum int,
	visit func(xml.StartElement, int)) (xml.Name, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	root := xml.Name{}
	depth := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if depth != 0 || root.Local == "" {
				return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
			}
			return root, nil
		}
		if err != nil {
			return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		switch value := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				if root.Local != "" {
					return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
				}
				root = value.Name
			}
			depth++
			*count++
			if depth > 128 || *count > maximum {
				return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
			}
			if visit != nil {
				visit(value, depth)
			}
		case xml.EndElement:
			depth--
			if depth < 0 {
				return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
			}
		case xml.Directive:
			return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(value)) != 0 {
				return xml.Name{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
			}
		}
	}
}

func (v *OfficeContentVerifier) checkPackage(data []byte, extension string) error {
	if !completeOfficeZIP(data) {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(archive.File) < 1 || len(archive.File) > v.maximumParts {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	parts := make(map[string][]byte, len(archive.File))
	seen := make(map[string]bool, len(archive.File))
	decoded := int64(0)
	elements := 0
	for _, part := range archive.File {
		if !officePartName(part.Name) || part.FileInfo().IsDir() || part.Flags&1 != 0 ||
			(part.Method != zip.Store && part.Method != zip.Deflate) || seen[part.Name] ||
			part.UncompressedSize64 > uint64(v.maximumDecodedBytes-decoded) {
			return uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		seen[part.Name] = true
		reader, err := part.Open()
		if err != nil {
			return uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, v.maximumDecodedBytes-decoded+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || int64(len(content)) != int64(part.UncompressedSize64) ||
			int64(len(content)) > v.maximumDecodedBytes-decoded {
			return uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		decoded += int64(len(content))
		if strings.HasSuffix(part.Name, ".xml") || strings.HasSuffix(part.Name, ".rels") {
			if _, err := scanOfficeXML(content, &elements, v.maximumElements, nil); err != nil {
				return err
			}
		}
		parts[part.Name] = content
	}
	main := "word/document.xml"
	contentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	if extension == "xlsx" {
		main = "xl/workbook.xml"
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	}
	if parts["[Content_Types].xml"] == nil || parts["_rels/.rels"] == nil || parts[main] == nil {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	typesMatch := false
	root, err := scanOfficeXML(parts["[Content_Types].xml"], new(int), v.maximumElements,
		func(element xml.StartElement, depth int) {
			if depth == 2 && element.Name.Space == officePackageNamespace+"content-types" && element.Name.Local == "Override" &&
				officeAttribute(element, "PartName", "") == "/"+main &&
				officeAttribute(element, "ContentType", "") == contentType {
				typesMatch = true
			}
		})
	if err != nil || root.Space != officePackageNamespace+"content-types" || root.Local != "Types" || !typesMatch {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	relationMatch := false
	root, err = scanOfficeXML(parts["_rels/.rels"], new(int), v.maximumElements,
		func(element xml.StartElement, depth int) {
			if depth == 2 && element.Name.Space == officePackageNamespace+"relationships" && element.Name.Local == "Relationship" &&
				officeAttribute(element, "Type", "") == officeDocumentNamespace+"relationships/officeDocument" &&
				officeAttribute(element, "Target", "") == main &&
				officeAttribute(element, "TargetMode", "") != "External" {
				relationMatch = true
			}
		})
	if err != nil || root.Space != officePackageNamespace+"relationships" || root.Local != "Relationships" || !relationMatch {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if extension == "docx" {
		body := false
		root, err = scanOfficeXML(parts[main], new(int), v.maximumElements,
			func(element xml.StartElement, depth int) {
				if depth == 2 && element.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" && element.Name.Local == "body" {
					body = true
				}
			})
		if err != nil || root.Space != "http://schemas.openxmlformats.org/wordprocessingml/2006/main" || root.Local != "document" || !body {
			return uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		return nil
	}
	workbookNamespace := "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	relationshipNamespace := officeDocumentNamespace + "relationships"
	sheetIDs := make(map[string]bool)
	root, err = scanOfficeXML(parts[main], new(int), v.maximumElements,
		func(element xml.StartElement, depth int) {
			if depth == 3 && element.Name.Space == workbookNamespace && element.Name.Local == "sheet" {
				id := officeAttribute(element, "id", relationshipNamespace)
				if id != "" {
					sheetIDs[id] = true
				}
			}
		})
	if err != nil || root.Space != workbookNamespace || root.Local != "workbook" || len(sheetIDs) == 0 {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	workbookRels := parts["xl/_rels/workbook.xml.rels"]
	if workbookRels == nil {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	matched := make(map[string]bool)
	root, err = scanOfficeXML(workbookRels, new(int), v.maximumElements,
		func(element xml.StartElement, depth int) {
			if depth != 2 || element.Name.Space != officePackageNamespace+"relationships" || element.Name.Local != "Relationship" ||
				officeAttribute(element, "Type", "") != officeDocumentNamespace+"relationships/worksheet" ||
				officeAttribute(element, "TargetMode", "") == "External" {
				return
			}
			id := officeAttribute(element, "Id", "")
			target := officeAttribute(element, "Target", "")
			if !sheetIDs[id] || !officePartName(target) {
				return
			}
			resolved := path.Clean(path.Join("xl", target))
			if !strings.HasPrefix(resolved, "xl/worksheets/") || parts[resolved] == nil {
				return
			}
			worksheetRoot, worksheetErr := scanOfficeXML(parts[resolved], new(int), v.maximumElements, nil)
			if worksheetErr == nil && worksheetRoot.Space == workbookNamespace && worksheetRoot.Local == "worksheet" {
				matched[id] = true
			}
		})
	if err != nil || root.Space != officePackageNamespace+"relationships" || root.Local != "Relationships" || len(matched) != len(sheetIDs) {
		return uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return nil
}

func (v *OfficeContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	if extension != "docx" && extension != "xlsx" {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if v == nil || source == nil || maximumBytes < 1 {
		return VerifiedContent{}, uploadError("BLOB_POLICY_INVALID")
	}
	select {
	case v.permits <- struct{}{}:
		defer func() { <-v.permits }()
	default:
		return VerifiedContent{}, uploadError("BLOB_VERIFICATION_BUSY")
	}
	bound := maximumBytes
	if bound > v.maximumInputBytes {
		bound = v.maximumInputBytes
	}
	data, err := io.ReadAll(io.LimitReader(source, bound+1))
	if err != nil {
		return VerifiedContent{}, uploadError("BLOB_STORE_FAILED")
	}
	if int64(len(data)) > bound {
		return VerifiedContent{}, uploadError("BLOB_SIZE_EXCEEDED")
	}
	if err := v.checkPackage(data, extension); err != nil {
		return VerifiedContent{}, err
	}
	mime := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	if extension == "xlsx" {
		mime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	digest := sha256.Sum256(data)
	return VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: mime}, nil
}
