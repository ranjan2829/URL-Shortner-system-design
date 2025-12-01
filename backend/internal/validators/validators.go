package validators

import (
	"net/url"
	"regexp"
)

func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Simple regex for basic URL validation if needed, but url.Parse is usually better
var urlRegex = regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(?:/[^/]*)*$`)

func IsValidURLRegex(str string) bool {
	return urlRegex.MatchString(str)
}
