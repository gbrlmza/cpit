package cpit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	// ReziseMode constants
	ResizeModeThumbnail   = "thumbnail"
	ResizeModeBestFit     = "bestFit"
	ResizeModeResize      = "resize"
	ResizeModeFitToWidth  = "fitToWidth"
	ResizeModeFitToHeight = "fitToHeight"

	// MimeType constants
	MimeTypeAuto = "auto"
	MimeTypeGif  = "gif"
	MimeTypeJpeg = "jpeg"
	MimeTypePng  = "png"
	MimeTypeWebp = "webp"
	MimeTypeBmp  = "bmp"

	// Cockpit API paths
	pathGetImage     = "/assets/image/%s"
	pathGetAsset     = "/assets/%s"
	pathGetSingleton = "/content/item/%s"
	pathUpsertItem   = "/content/item/%s"
	pathGetItem      = "/content/item/%s/%s"
	pathDeleteItem   = "/content/item/%s/%s"
	pathGetItems     = "/content/items/%s"
)

type (
	// optionFn is a function that applies an option to a cockpitReq
	optionFn func(r *cockpitReq) error

	// cockpitReq is a request to the cockpit API
	cockpitReq struct {
		httpClient   *http.Client
		apiKey       string
		baseURL      string
		method       string
		path         string
		params       url.Values
		body         interface{}
		debug        bool
		output       interface{}
		outputHeader *http.Header
		isImage      bool
	}
)

var (
	// defaultHttpClient is the default http client used for requests
	defaultHttpClient *http.Client
	// defaultBaseUrl is the default base url used for requests
	defaultBaseUrl string
	// defaultApiKey is the default api key used for requests
	defaultApiKey string
	// defaultDebugMode is the default debug mode used for requests
	defaultDebugMode bool
	// mtx is a mutex to protect the default values
	mtx = &sync.RWMutex{}
)

// SetDefaultHttpClient sets the default http client used for requests.
func SetDefaultHttpClient(c *http.Client) {
	mtx.Lock()
	defer mtx.Unlock()
	defaultHttpClient = c
}

// SetDefaultBaseUrl sets the default base url used for requests.
func SetDefaultBaseUrl(url string) {
	mtx.Lock()
	defer mtx.Unlock()
	defaultBaseUrl = url
}

// SetDefaultApiKey sets the default api key used for requests.
func SetDefaultApiKey(key string) {
	mtx.Lock()
	defer mtx.Unlock()
	defaultApiKey = key
}

// SetDefaultDebugMode sets the default debug mode used for requests.
func SetDefaultDebugMode(enabled bool) {
	mtx.Lock()
	defer mtx.Unlock()
	defaultDebugMode = enabled
}

// newCockpitReq returns a new cockpitReq with the default values.
func newCockpitReq() cockpitReq {
	mtx.RLock()
	defer mtx.RUnlock()
	return cockpitReq{
		httpClient: defaultHttpClient,
		apiKey:     defaultApiKey,
		baseURL:    defaultBaseUrl,
		params:     make(url.Values),
		debug:      defaultDebugMode,
	}
}

// GetItems requests items of a model.
// - model is the name of the model
func GetItems(ctx context.Context, model string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodGet
	r.path = fmt.Sprintf(pathGetItems, model)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// GetSingleton requests a singleton model.
// - model is the name of the model
func GetSingleton(ctx context.Context, model string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodGet
	r.path = fmt.Sprintf(pathGetSingleton, model)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// GetAsset requests an asset.
// - id is the id of the asset
func GetAsset(ctx context.Context, id string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodGet
	r.path = fmt.Sprintf(pathGetAsset, id)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// GetImage requests an image.
// - id is the id of the image
func GetImage(ctx context.Context, id string, opts ...optionFn) (string, error) {
	r := newCockpitReq()
	r.method = http.MethodGet
	r.path = fmt.Sprintf(pathGetImage, id)
	r.isImage = true
	if err := applyOptions(&r, opts); err != nil {
		return "", err
	}
	return r.runImage(ctx)
}

// GetItem requests an item.
// - model is the name of the model
// - id is the id of the item
func GetItem(ctx context.Context, model string, id string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodGet
	r.path = fmt.Sprintf(pathGetItem, model, id)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// UpsertItem upserts an item.
// - model is the name of the model
func UpsertItem(ctx context.Context, model string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodPost
	r.path = fmt.Sprintf(pathUpsertItem, model)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// DeleteItem deletes an item.
// - model is the name of the model
// - id is the id of the item
func DeleteItem(ctx context.Context, model string, id string, opts ...optionFn) error {
	r := newCockpitReq()
	r.method = http.MethodDelete
	r.path = fmt.Sprintf(pathDeleteItem, model, id)
	if err := applyOptions(&r, opts); err != nil {
		return err
	}
	return r.run(ctx)
}

// applyOptions applies the options to the request and executes it.
func applyOptions(r *cockpitReq, opts []optionFn) error {
	for _, o := range opts {
		if err := o(r); err != nil {
			return err
		}
	}

	return nil
}

// WithHttpClient sets the http client for the request.
func WithHttpClient(c *http.Client) optionFn {
	return func(r *cockpitReq) error {
		r.httpClient = c
		return nil
	}
}

// WithBaseURL sets the base url for the request.
func WithBaseURL(url string) optionFn {
	return func(r *cockpitReq) error {
		if url == "" {
			return errors.New("url is required")
		}

		r.baseURL = url
		return nil
	}
}

// WithApiKey sets the api key for the request.
func WithApiKey(key string) optionFn {
	return func(r *cockpitReq) error {
		if key == "" {
			return errors.New("apikey is required")
		}

		r.apiKey = key
		return nil
	}
}

// WithDebugMode sets the debug mode for the request.
func WithDebugMode(enabled bool) optionFn {
	return func(r *cockpitReq) error {
		r.debug = enabled
		return nil
	}
}

// WithOutputHeader sets where the response headers should be written to.
func WithOutputHeader(h *http.Header) optionFn {
	return func(r *cockpitReq) error {
		r.outputHeader = h
		return nil
	}
}

// WithOutput sets where the output should be written to.
// The response body will be decoded as json into the output.
func WithOutput(o interface{}) optionFn {
	return func(r *cockpitReq) error {
		if o == nil || reflect.TypeOf(o).Kind() != reflect.Ptr {
			return errors.New("output must be a pointer")
		}
		r.output = o
		return nil
	}
}

// WithBody sets the body for the request.
// The body can be an io.Reader, []byte, string or any other type that can be encoded as json.
func WithBody(b interface{}) optionFn {
	return func(r *cockpitReq) error {
		r.body = b
		return nil
	}
}

// WithResizeMode sets the resize mode for the image.
func WithResizeMode(m string) optionFn {
	return func(r *cockpitReq) error {
		if m != ResizeModeThumbnail &&
			m != ResizeModeBestFit &&
			m != ResizeModeResize &&
			m != ResizeModeFitToWidth &&
			m != ResizeModeFitToHeight {
			return errors.New("invalid resize mode")
		}
		r.params.Set("m", m)
		return nil
	}
}

// WithWidth sets the width for the image.
func WithWidth(w int) optionFn {
	return func(r *cockpitReq) error {
		if w < 1 {
			return errors.New("width must be greater than 0")
		}
		r.params.Set("w", strconv.Itoa(w))
		return nil
	}
}

// WithHeight sets the height for the image.
func WithHeight(h int) optionFn {
	return func(r *cockpitReq) error {
		if h < 1 {
			return errors.New("height must be greater than 0")
		}
		r.params.Set("h", strconv.Itoa(h))
		return nil
	}
}

// WithQuality sets the quality for the image.
func WithQuality(q int) optionFn {
	return func(r *cockpitReq) error {
		if q < 1 || q > 100 {
			return errors.New("quality must be between 1 and 100")
		}
		r.params.Set("q", strconv.Itoa(q))
		return nil
	}
}

// WithMime sets the mime type for the image.
func WithMime(mime string) optionFn {
	return func(r *cockpitReq) error {
		if mime != MimeTypeAuto &&
			mime != MimeTypeGif &&
			mime != MimeTypeJpeg &&
			mime != MimeTypePng &&
			mime != MimeTypeWebp &&
			mime != MimeTypeBmp {
			return errors.New("invalid mime type")
		}

		r.params.Set("mime", mime)
		return nil
	}
}

// WithLocale specifies the wanted locale.
func WithLocale(locale string) optionFn {
	return func(r *cockpitReq) error {
		r.params.Set("locale", locale)
		return nil
	}
}

// WithFields specifies the fields to be returned in the request. See following links for more info:
// https://getcockpit.com/documentation/core/api/content#project-fields-to-return-from-query
// https://www.mongodb.com/docs/manual/tutorial/project-fields-from-query-results/
// Example to only retrieve the title field:
// - WithFields(`{"title":1}`)
func WithFields(fields string) optionFn {
	return func(r *cockpitReq) error {
		r.params.Set("fields", fields)
		return nil
	}
}

// WithFilter specifies the filter to be applied in the request. See following links for more info:
// https://getcockpit.com/documentation/core/api/content#filtering
// https://www.mongodb.com/docs/manual/reference/operator/query/
// Example to filter items with title containing "cat" (case insensitive):
// - WithFilter(`{"title": {"$regex": "/cat/i"}}`)
func WithFilter(filter string) optionFn {
	return func(r *cockpitReq) error {
		r.params.Set("filter", filter)
		return nil
	}
}

// WithSort specifies the sort order for the request.
// Example to sort by title in ascending order:
// - WithSort(`{"title":1}`)
// Example to sort by title in descending order:
// - WithSort(`{"title":-1}`)
func WithSort(sort string) optionFn {
	return func(r *cockpitReq) error {
		r.params.Set("sort", sort)
		return nil
	}
}

// WithLimit specifies the number of items to return.
func WithLimit(limit int) optionFn {
	return func(r *cockpitReq) error {
		if limit < 1 {
			return errors.New("limit must be greater than 0")
		}
		r.params.Set("limit", strconv.Itoa(limit))
		return nil
	}
}

// WithSkip specifies the number of items to skip. Useful for pagination, it
// must be used in combination with WithLimit or it won't have any effect. When
// using pagination(WithSkip), cockit uses a different response format. The
// PaginatedResp model can be used as output to get the total number of items.
func WithSkip(skip int) optionFn {
	return func(r *cockpitReq) error {
		if skip < 0 {
			return errors.New("skip must be greater than or equal to 0")
		}
		r.params.Set("skip", strconv.Itoa(skip))
		return nil
	}
}

// WithPopulate specifies if linked content items should be populated.
func WithPopulate(enabled bool) optionFn {
	return func(r *cockpitReq) error {
		v := "0"
		if enabled {
			v = "1"
		}

		r.params.Set("populate", v)
		return nil
	}
}

// run executes the request and parses the response.
func (r *cockpitReq) run(ctx context.Context) error {
	resp, err := r.doHttp(ctx)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), b)
	}

	if r.output != nil {
		if err := json.NewDecoder(resp.Body).Decode(r.output); err != nil {
			return err
		}
	}

	if r.outputHeader != nil {
		*r.outputHeader = resp.Header
	}

	return nil
}

// run executes the request and parses the response as an image url.
func (r *cockpitReq) runImage(ctx context.Context) (string, error) {
	resp, err := r.doHttp(ctx)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), body)
	}

	if r.outputHeader != nil {
		*r.outputHeader = resp.Header
	}

	return string(body), nil
}

// doHttp executes the http request and returns the response.
func (r *cockpitReq) doHttp(ctx context.Context) (*http.Response, error) {
	if r.apiKey == "" {
		return nil, errors.New("apiKey is required. either set it as default or pass it as an option")
	}
	if r.baseURL == "" {
		return nil, errors.New("baseURL is required. either set it as default or pass it as an option")
	}

	httpClient := r.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	body, err := getBodyReader(r.body)
	if err != nil {
		return nil, err
	}

	if r.method == http.MethodPost && body == nil {
		return nil, errors.New("body is required for upserting an item")
	}

	url := fmt.Sprintf("%s%s?%s", r.baseURL, r.path, r.params.Encode())
	req, err := http.NewRequestWithContext(ctx, r.method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("api-key", r.apiKey)

	resp, err := httpClient.Do(req)

	if r.debug {
		// log request url and response if debug mode is enabled. the api-key header is not logged
		var body []byte
		var sc string
		if resp != nil {
			body, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(body))
			sc = fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		if err != nil {
			sc = "ERROR"
			body = []byte(err.Error())
		}
		slog.Debug(fmt.Sprintf("[Cockpit][%s] %s %s | %s\n", sc, r.method, url, body))
	}

	return resp, err
}

// getBodyReader returns an io.Reader for the given body
func getBodyReader(body interface{}) (io.Reader, error) {
	var r io.Reader

	switch b := body.(type) {
	case nil:
	case io.Reader:
		r = b
	case []byte:
		r = bytes.NewReader(b)
	case string:
		r = strings.NewReader(b)
	default:
		// try to encode it as json
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode body: %w", err)
		}
		r = buf
	}

	return r, nil
}
