package structs

import (
	"net/http"
)

type TargetUrl struct {
	Url      string         `json:"url"`
	UrlProto string         `json:"urlproto"`
	Domain   string         `json:"domain"`
	Port     string         `json:"port"`
	Code     int            `json:"code"`
	Protocol string         `json:"protocol"`
	Cookies  []*http.Cookie `json:"cookies"`
	Headers  http.Header    `json:"headers"`
	Body     string         `json:"body"`
}
