// Package client provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.5.0 DO NOT EDIT.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime"
)

// Defines values for CheckReportStatus.
const (
	CheckReportStatusCompleted CheckReportStatus = "completed"
	CheckReportStatusDisabled  CheckReportStatus = "disabled"
	CheckReportStatusError     CheckReportStatus = "error"
	CheckReportStatusFail      CheckReportStatus = "fail"
	CheckReportStatusPass      CheckReportStatus = "pass"
	CheckReportStatusSkipped   CheckReportStatus = "skipped"
	CheckReportStatusUnknown   CheckReportStatus = "unknown"
)

// Defines values for GetComponentReportsParamsStatus.
const (
	GetComponentReportsParamsStatusCompleted GetComponentReportsParamsStatus = "completed"
	GetComponentReportsParamsStatusDisabled  GetComponentReportsParamsStatus = "disabled"
	GetComponentReportsParamsStatusError     GetComponentReportsParamsStatus = "error"
	GetComponentReportsParamsStatusFail      GetComponentReportsParamsStatus = "fail"
	GetComponentReportsParamsStatusPass      GetComponentReportsParamsStatus = "pass"
	GetComponentReportsParamsStatusSkipped   GetComponentReportsParamsStatus = "skipped"
	GetComponentReportsParamsStatusUnknown   GetComponentReportsParamsStatus = "unknown"
)

// CheckReport A quality check report for a component
type CheckReport struct {
	// CheckSlug Unique identifier for the check type
	CheckSlug string `json:"check_slug"`

	// Id Unique identifier for the report
	Id string `json:"id"`

	// Status Status of the check execution
	Status CheckReportStatus `json:"status"`

	// Timestamp When the check was executed
	Timestamp time.Time `json:"timestamp"`
}

// CheckReportStatus Status of the check execution
type CheckReportStatus string

// Component A component discovered from a source
type Component struct {
	// Description Additional context about the component's purpose and functionality
	Description *string `json:"description,omitempty"`

	// Id Unique identifier for the component. If not provided, the name will be used as the identifier.
	Id *string `json:"id,omitempty"`

	// Name Human-readable name of the component
	Name string `json:"name"`

	// Owners Ownership information for a component
	Owners *Owners `json:"owners,omitempty"`
}

// ComponentReportsResponse Response containing component reports with pagination
type ComponentReportsResponse struct {
	// Pagination Pagination metadata for list responses
	Pagination Pagination `json:"pagination"`

	// Reports List of check reports for the component
	Reports []CheckReport `json:"reports"`
}

// Error Error response
type Error struct {
	// Code Error code
	Code *string `json:"code,omitempty"`

	// Error Error message
	Error string `json:"error"`
}

// Owners Ownership information for a component
type Owners struct {
	// Maintainers List of user identifiers responsible for maintaining this component
	Maintainers *[]string `json:"maintainers,omitempty"`

	// Team Team responsible for owning this component
	Team *string `json:"team,omitempty"`
}

// Pagination Pagination metadata for list responses
type Pagination struct {
	// HasMore Whether there are more items available
	HasMore bool `json:"has_more"`

	// Limit Number of items returned in this response
	Limit int `json:"limit"`

	// Offset Offset used for this response
	Offset int `json:"offset"`

	// Total Total number of items available
	Total int `json:"total"`
}

// GetComponentReportsParams defines parameters for GetComponentReports.
type GetComponentReportsParams struct {
	// Status Filter by check status
	Status *GetComponentReportsParamsStatus `form:"status,omitempty" json:"status,omitempty"`

	// CheckSlug Filter by specific check type
	CheckSlug *string `form:"check_slug,omitempty" json:"check_slug,omitempty"`

	// Since Filter reports since timestamp (ISO 8601)
	Since *time.Time `form:"since,omitempty" json:"since,omitempty"`

	// Limit Number of reports to return
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`

	// Offset Pagination offset
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`

	// LatestPerCheck Return only the latest report for each check type
	LatestPerCheck *bool `form:"latest_per_check,omitempty" json:"latest_per_check,omitempty"`
}

// GetComponentReportsParamsStatus defines parameters for GetComponentReports.
type GetComponentReportsParamsStatus string

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetComponents request
	GetComponents(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetComponentById request
	GetComponentById(ctx context.Context, componentId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetComponentReports request
	GetComponentReports(ctx context.Context, componentId string, params *GetComponentReportsParams, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetComponents(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetComponentsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetComponentById(ctx context.Context, componentId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetComponentByIdRequest(c.Server, componentId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetComponentReports(ctx context.Context, componentId string, params *GetComponentReportsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetComponentReportsRequest(c.Server, componentId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetComponentsRequest generates requests for GetComponents
func NewGetComponentsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/catalog/v1/components")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetComponentByIdRequest generates requests for GetComponentById
func NewGetComponentByIdRequest(server string, componentId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "componentId", runtime.ParamLocationPath, componentId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/catalog/v1/components/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetComponentReportsRequest generates requests for GetComponentReports
func NewGetComponentReportsRequest(server string, componentId string, params *GetComponentReportsParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "componentId", runtime.ParamLocationPath, componentId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/catalog/v1/components/%s/reports", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.Status != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "status", runtime.ParamLocationQuery, *params.Status); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.CheckSlug != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "check_slug", runtime.ParamLocationQuery, *params.CheckSlug); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.Since != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "since", runtime.ParamLocationQuery, *params.Since); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.Limit != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "limit", runtime.ParamLocationQuery, *params.Limit); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.Offset != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "offset", runtime.ParamLocationQuery, *params.Offset); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.LatestPerCheck != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "latest_per_check", runtime.ParamLocationQuery, *params.LatestPerCheck); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetComponentsWithResponse request
	GetComponentsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetComponentsResponse, error)

	// GetComponentByIdWithResponse request
	GetComponentByIdWithResponse(ctx context.Context, componentId string, reqEditors ...RequestEditorFn) (*GetComponentByIdResponse, error)

	// GetComponentReportsWithResponse request
	GetComponentReportsWithResponse(ctx context.Context, componentId string, params *GetComponentReportsParams, reqEditors ...RequestEditorFn) (*GetComponentReportsResponse, error)
}

type GetComponentsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Component
	JSON500      *Error
}

// Status returns HTTPResponse.Status
func (r GetComponentsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetComponentsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetComponentByIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Component
	JSON404      *Error
	JSON500      *Error
}

// Status returns HTTPResponse.Status
func (r GetComponentByIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetComponentByIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetComponentReportsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ComponentReportsResponse
	JSON404      *Error
	JSON500      *Error
}

// Status returns HTTPResponse.Status
func (r GetComponentReportsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetComponentReportsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetComponentsWithResponse request returning *GetComponentsResponse
func (c *ClientWithResponses) GetComponentsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetComponentsResponse, error) {
	rsp, err := c.GetComponents(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetComponentsResponse(rsp)
}

// GetComponentByIdWithResponse request returning *GetComponentByIdResponse
func (c *ClientWithResponses) GetComponentByIdWithResponse(ctx context.Context, componentId string, reqEditors ...RequestEditorFn) (*GetComponentByIdResponse, error) {
	rsp, err := c.GetComponentById(ctx, componentId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetComponentByIdResponse(rsp)
}

// GetComponentReportsWithResponse request returning *GetComponentReportsResponse
func (c *ClientWithResponses) GetComponentReportsWithResponse(ctx context.Context, componentId string, params *GetComponentReportsParams, reqEditors ...RequestEditorFn) (*GetComponentReportsResponse, error) {
	rsp, err := c.GetComponentReports(ctx, componentId, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetComponentReportsResponse(rsp)
}

// ParseGetComponentsResponse parses an HTTP response from a GetComponentsWithResponse call
func ParseGetComponentsResponse(rsp *http.Response) (*GetComponentsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetComponentsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Component
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseGetComponentByIdResponse parses an HTTP response from a GetComponentByIdWithResponse call
func ParseGetComponentByIdResponse(rsp *http.Response) (*GetComponentByIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetComponentByIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Component
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseGetComponentReportsResponse parses an HTTP response from a GetComponentReportsWithResponse call
func ParseGetComponentReportsResponse(rsp *http.Response) (*GetComponentReportsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetComponentReportsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ComponentReportsResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}
