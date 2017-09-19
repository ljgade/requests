package requests

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Request interface {
	SetHeader(string, ...string) Request
	Headers() map[string][]string

	SetQueryParam(string, ...string) Request
	QueryParams() map[string][]string
	UrlPath() string
	SetJSON(string) Request
	SetRawBody([]byte) Request

	SetFormParam(string, ...string) Request
	FormParams() map[string][]string

	AddFile(string, string, []byte) Request

	Send() (Response, error)
}

type formFile struct {
	filename string
	data     []byte
}

type request struct {
	*session
	method  string
	URL     string
	headers map[string][]string

	isJSON      bool
	body        []byte
	formParams  map[string][]string
	queryParams map[string][]string
	files       map[string]*formFile
}

// Validate URLPath
func parseURL(urlPath string) (URL *url.URL, err error) {
	// First pass
	URL, err = url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	// To check the Scheme after the first pass, if it is neither http or https, then make it default to be http and validate again
	if URL.Scheme != "http" && URL.Scheme != "https" {
		urlPath = "http://" + urlPath
		URL, err = url.Parse(urlPath)
		if err != nil {
			return nil, err
		}

		// Accepts only http and https scheme
		if URL.Scheme != "http" && URL.Scheme != "https" {
			return nil, errors.New("[package requests] only HTTP and HTTPS are accepted")
		}
	}
	return
}

// Create a request
func newRequest(method string, urlPath string, s *session) (Request, error) {
	// Validate URLPath
	URL, err := parseURL(urlPath)
	if err != nil {
		return nil, err
	}

	// Extract the url params from the urlpath
	queryParams := make(map[string][]string)
	for key, values := range URL.Query() {
		queryParams[key] = values
	}

	urlPath = URL.Scheme + "://" + URL.Host + URL.Path
	r := &request{session: s, method: method, URL: urlPath}
	r.headers = make(map[string][]string)
	r.formParams = make(map[string][]string)
	r.queryParams = queryParams
	r.files = make(map[string]*formFile)
	return r, nil
}

// Set a request header, could be multiple values. If no values are provided, then delete the key if any.
func (this *request) SetHeader(key string, values ...string) Request {
	if len(values) > 0 {
		this.headers[key] = values[:]
	} else {
		delete(this.headers, key)
	}
	return this
}

// Get a copy of request headers, any modification made to this map WILL NOT reflect back to the actually request headers
func (this *request) Headers() map[string][]string {
	headers := make(map[string][]string)
	for key, values := range this.headers {
		headers[key] = values[:]
	}
	return headers
}

// Set a url param, could be multiple values. If no values are provided, then delete the key if any.
func (this *request) SetQueryParam(key string, values ...string) Request {
	if len(values) > 0 {
		this.queryParams[key] = values[:]
	} else {
		delete(this.queryParams, key)
	}
	return this
}

// Get a copy of url params, any modification made to this map WILL NOT reflect back to the actually url params
func (this *request) QueryParams() map[string][]string {
	params := make(map[string][]string)
	for key, values := range this.queryParams {
		params[key] = values[:]
	}
	return params
}

// Get the full url path
func (this *request) UrlPath() string {
	if len(this.queryParams) > 0 {
		return this.URL + "?" + parseParams(this.queryParams).Encode()
	} else {
		return this.URL
	}
}

// Set a JSON message(Content-Type header will be "application/json")
func (this *request) SetJSON(json string) Request {
	this.isJSON = true
	this.body = []byte(json)
	return this
}

// Set raw message body
// NOTICE: it is the users' responsability to set the correct Content-Type header
func (this *request) SetRawBody(body []byte) Request {
	this.isJSON = false
	this.body = body
	return this
}

// Set a body param, could be multiple values. If no values are provided, then delete the key if any.
func (this *request) SetFormParam(key string, values ...string) Request {
	if len(values) > 0 {
		this.formParams[key] = values[:]
	} else {
		delete(this.formParams, key)
	}
	return this
}

// Get a copy of body params, any modification made to this map WILL NOT reflect back to the actually body params
func (this *request) FormParams() map[string][]string {
	params := make(map[string][]string)
	for key, values := range this.queryParams {
		params[key] = values[:]
	}
	return params
}

// Add a file
func (this *request) AddFile(fieldname string, filename string, data []byte) Request {
	if fieldname != "" && filename != "" && data != nil {
		this.files[fieldname] = &formFile{filename: filename, data: data}
	}
	return this
}

func (this *request) parseBody() (req *http.Request, err error) {
	// GET and TRACE request should not have a message body
	if this.method == "GET" || this.method == "TRACE" {
		req, err = http.NewRequest(this.method, this.UrlPath(), nil)
	}

	// Process message body
	if len(this.body) > 0 {
		if this.isJSON {
			this.headers["Content-Type"] = []string{"application/json"}
			req, err = http.NewRequest(this.method, this.UrlPath(),
				strings.NewReader(string(this.body)))
		} else {
			var body *bytes.Buffer
			body = bytes.NewBuffer(this.body)
			req, err = http.NewRequest(this.method, this.UrlPath(), body)
		}
	} else if len(this.files) > 0 {
		// multipart
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		for fieldname, values := range this.formParams {
			err = writer.WriteField(fieldname, values[0])
			if err != nil {
				return
			}
		}
		var part io.Writer
		for fieldname, file := range this.files {
			part, err = writer.CreateFormFile(fieldname, file.filename)
			if err != nil {
				return
			}
			_, err = part.Write(file.data)
			if err != nil {
				return
			}
		}
		err = writer.Close()
		if err != nil {
			return
		}
		this.headers["Content-Type"] = []string{writer.FormDataContentType()}
		req, err = http.NewRequest(this.method, this.UrlPath(), body)
	} else {
		this.headers["Content-Type"] = []string{"application/x-www-form-urlencoded"}
		req, err = http.NewRequest(this.method, this.UrlPath(),
			strings.NewReader(parseParams(this.formParams).Encode()))
	}
	return
}

func (this *request) Send() (res Response, err error) {
	req, err := this.parseBody()
	if err != nil {
		return
	}
	this.session.setCookies(req.URL)
	req.Header = parseHeaders(this.headers)
	httpResponse, err := this.session.Do(req)
	if err != nil {
		return
	}
	res, err = newResponse(httpResponse)
	return
}

func (this *request) SendRequestWithoutParseBody(httpRequest *http.Request) (res Response, err error) {
	req, err := http.NewRequest(this.method, this.UrlPath(), httpRequest.Body)
	if err != nil {
		return
	}
	this.session.setCookies(req.URL)
	req.Header = parseHeaders(this.headers)
	httpResponse, err := this.session.Do(req)
	if err != nil {
		return
	}
	res, err = newResponse(httpResponse)
	return
}
