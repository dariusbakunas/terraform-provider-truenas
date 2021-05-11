//	TrueNAS client code is based on go github client,
//	github copyright notice is as follows:
//
//	Copyright (c) 2013 The go-github AUTHORS. All rights reserved.
//
//	Redistribution and use in source and binary forms, with or without
//	modification, are permitted provided that the following conditions are
//	met:
//
//	  * Redistributions of source code must retain the above copyright
//	notice, this list of conditions and the following disclaimer.
//	  * Redistributions in binary form must reproduce the above
//	copyright notice, this list of conditions and the following disclaimer
//	in the documentation and/or other materials provided with the
//	distribution.
//	  * Neither the name of Google Inc. nor the names of its
//	contributors may be used to endorse or promote products derived from
//	this software without specific prior written permission.
//
//	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
//	"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
//	LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
//	A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//	OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//	SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
//	LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
//	DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
//	THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//	(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
//	OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	userAgent = "go-terraform-provider-truenas"
)

type service struct {
	client *Client
}

type ListOptions struct {
	Limit  int    `url:"limit,omitempty"`
	Offset int    `url:"offset,omitempty"`
	Count  bool   `url:"count,omitempty"`
	Sort   string `url:"sort,omitempty"`
}

type Client struct {
	client *http.Client // HTTP client used to communicate with the API

	// Base URL for TrueNAS API requests.
	BaseURL *url.URL

	// User agent used when communicating with the TrueNAS API.
	UserAgent string

	// Services used for talking to different parts of the TrueNAS API.
	PoolAPI    *PoolService
	DatasetAPI *DatasetService
	ServiceAPI *ServiceAPI
}

// NewClient returns a new TrueNAS API client.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	baseEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c := &Client{client: httpClient, BaseURL: baseEndpoint, UserAgent: userAgent}

	baseService := service{c}
	c.PoolAPI = (*PoolService)(&baseService)
	c.DatasetAPI = (*DatasetService)(&baseService)
	c.ServiceAPI = (*ServiceAPI)(&baseService)

	return c, nil
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c < 300 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Message = string(data)
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	// TODO: handle more status codes here

	return errorResponse
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if errURL, err := url.Parse(e.URL); err == nil {
				e.URL = errURL.String()
				return nil, e
			}
		}

		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)

	if err != nil {
		return resp, err
	}

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}

	return resp, err
}
