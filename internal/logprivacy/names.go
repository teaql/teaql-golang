package logprivacy

import "strings"

// CredentialName is a conservative fail-closed classifier shared by SQL and audit.
func CredentialName(name string) bool {
	name = strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, name)
	for _, word := range []string{"password", "passwd", "passphrase", "privatekey", "secret", "accesstoken", "refreshtoken", "idtoken", "apikey", "authorization", "credential", "sessiontoken", "magiclinktoken"} {
		if strings.Contains(name, word) {
			return true
		}
	}
	return false
}
