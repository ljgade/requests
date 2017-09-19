package requests

import (
	"errors"
	"net/http"
	"net/url"
	"crypto/tls"
)

const (
	Version = "1.0"
)

var (
	sessionWithoutCookies *session
)

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	s := &session{Client: client}
	s.cookies = make(map[string]string)
	sessionWithoutCookies = s
}

// Parse url params or body params. Usually the callers intend to get application/x-www-form-urlencoded format of the params
func parseParams(params map[string][]string) url.Values {
	v := url.Values{}
	for key, values := range params {
		for _, value := range values {
			v.Add(key, value)
		}
	}
	return v
}

// Parse the headers, with some default values added
func parseHeaders(headers map[string][]string) http.Header {
	h := http.Header{}
	for key, values := range headers {
		for _, value := range values {
			h.Add(key, value)
		}
	}
	_, hasAccept := h["Accept"]
	if !hasAccept {
		h.Add("Accept", "*/*")
	}
	_, hasAgent := h["User-Agent"]
	if !hasAgent {
		h.Add("User-Agent", "go-requests/"+Version)
	}
	return h
}

// Thread-safe version implementations of the seven HTTP methods, but also do not have a cookiejar
func Method(method string, urlPath string) (Request, error) {
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" &&
		method != "HEAD" && method != "OPTIONS" && method != "PATCH" {
		return nil, errors.New("method not supported")
	}
	return newRequest(method, urlPath, sessionWithoutCookies)
}

func Get(urlPath string) (Request, error) {
	return newRequest("GET", urlPath, sessionWithoutCookies)
}

func Post(urlPath string) (Request, error) {
	return newRequest("POST", urlPath, sessionWithoutCookies)
}

func Put(urlPath string) (Request, error) {
	return newRequest("PUT", urlPath, sessionWithoutCookies)
}

func Delete(urlPath string) (Request, error) {
	return newRequest("DELETE", urlPath, sessionWithoutCookies)
}

func Head(urlPath string) (Request, error) {
	return newRequest("HEAD", urlPath, sessionWithoutCookies)
}

func Options(urlPath string) (Request, error) {
	return newRequest("OPTIONS", urlPath, sessionWithoutCookies)
}

func Patch(urlPath string) (Request, error) {
	return newRequest("PATCH", urlPath, sessionWithoutCookies)
}
