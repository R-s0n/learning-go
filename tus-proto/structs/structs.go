package structs

import (
	"net/http"
)

type TargetUrl struct {
	Code     int
	Protocol string
	Cookies  []*http.Cookie
	Headers  http.Header
	Body     string
}
