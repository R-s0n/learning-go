package structs

import (
	"net/http"
)

type TargetUrl struct {
	Code     int            `json:"code"`
	Protocol string         `json:"protocol"`
	Cookies  []*http.Cookie `json:"cookies"`
	Headers  http.Header    `json:"headers"`
	Body     string         `json:"body"`
}
