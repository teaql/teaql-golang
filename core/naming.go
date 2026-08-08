package core

import "strings"

func DefaultTableName(entityName string) string {
	var out strings.Builder
	out.Grow(len(entityName) + 5)

	for i, ch := range entityName {
		if ch >= 'A' && ch <= 'Z' {
			if i > 0 {
				out.WriteByte('_')
			}
			out.WriteByte(byte(ch + 32)) // to lower
		} else {
			out.WriteRune(ch)
		}
	}
	out.WriteString("_data")
	return out.String()
}
