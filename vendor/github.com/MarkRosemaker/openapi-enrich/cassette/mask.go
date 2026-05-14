package cassette

import (
	"strings"
)

// Mask redacts sensitive values in recorded request headers in place.
func (ias Interactions) Mask() {
	for i, ia := range ias {
		for k, vals := range ia.Request.Headers {
			if len(vals) == 0 {
				continue
			}

			switch k {
			case "Authorization":
				ias[i].Request.Headers[k] = []string{maskBearerToken(vals[0])}
			}
		}
	}
}

func maskBearerToken(token string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(token, prefix) {
		return token
	}

	return prefix + strings.Repeat("*", len(token)-len(prefix))
}
