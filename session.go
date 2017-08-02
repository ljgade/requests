package requests

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type Session interface {
	// TODO: implement timeout
	// SetTimeout(time.Duration) Session
	// Timeout() time.Duration
	SetCookies(map[string]string) Session
	Cookies() map[string]string

	Method(string, string) (Request, error)
	Get(string) (Request, error)
	Post(string) (Request, error)
	Put(string) (Request, error)
	Delete(string) (Request, error)
	Head(string) (Request, error)
	Options(string) (Request, error)
	Patch(string) (Request, error)
}

type session struct {
	*http.Client
	cookies map[string]string
}

// Create a session
func NewSession() Session {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	s := &session{Client: client}
	s.cookies = make(map[string]string)
	return s
}

func (this *session) SetTimeout(timeout time.Duration) Session {
	this.Client.Timeout = timeout
	return this
}

func (this *session) Timeout() time.Duration {
	return this.Client.Timeout
}

// initialize cookies for this session
func (this *session) SetCookies(cookies map[string]string) Session {
	for key, value := range cookies {
		this.cookies[key] = value
	}
	return this
}

// the cookies getter
func (this *session) Cookies() map[string]string {
	cookies := make(map[string]string)
	for key, value := range this.cookies {
		cookies[key] = value
	}
	return cookies
}

// should be called before every request
func (this *session) setCookies(URL *url.URL) {
	if this.Jar == nil {
		// this session is without a cookiejar
		return
	}

	cookies := this.Jar.Cookies(URL)
	for name, value := range this.cookies {
		// only sets the cookie when no corresponding one is found
		found := false
		for _, cookie := range cookies {
			if cookie.Name == name {
				found = true
				break
			}
		}
		if !found {
			cookies = append(cookies, &http.Cookie{Name: name, Value: value})
		}
	}
	this.Jar.SetCookies(URL, cookies)
}
