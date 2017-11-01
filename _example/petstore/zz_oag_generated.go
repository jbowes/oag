package petstore

import (
	bytes "bytes"
	context "context"
	json "encoding/json"
	http "net/http"
	url "net/url"
)

// This file is automatically generated by oag (https://github.com/jbowes/oag)
// DO NOT EDIT

const baseURL = "https://petstore.swagger.io/api"

// Pet is a data type for API communication.
type Pet struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Tag  *string `json:"tag"` // Optional
}

// PetIter Iterates over a result set of Pets.
type PetIter struct {
	page []Pet
	i    int

	err   error
	first bool
}

// Close closes the PetIter and releases any associated resources.
// After Close, any calls to Current will return an error.
func (i *PetIter) Close() {}

// Next advances the PetIter and returns a boolean indicating if the end has been reached.
// Next must be called before the first call to Current.
// Calls to Current after Next returns false will return an error.
func (i *PetIter) Next() bool {
	if i.first && i.err != nil {
		i.first = false
		return true
	}
	i.first = false
	i.i++
	return i.i < len(i.page)
}

// Current returns the current Pet, and an optional error. Once an error has been returned,
// the PetIter is closed, or the end of iteration is reached, subsequent calls to Current
// will return an error.
func (i *PetIter) Current() (*Pet, error) {
	if i.err != nil {
		return nil, i.err
	}
	return &i.page[i.i], nil
}

// PetsClient provides access to the /pets APIs
type PetsClient endpoint

// List corresponds to the GET /pets endpoint.
// Returns all pets from the system that the user has access to
func (c *PetsClient) List(ctx context.Context) *PetIter {
	iter := PetIter{
		first: true,
		i:     -1,
	}

	p := "/pets"

	var req *http.Request
	req, iter.err = c.backend.NewRequest(http.MethodGet, p, nil, nil)
	if iter.err != nil {
		return &iter
	}

	_, iter.err = c.backend.Do(ctx, req, &iter.page, nil)
	return &iter
}

// Backend defines the low-level interface for communicating with the remote api.
type Backend interface {
	NewRequest(method, path string, query url.Values, body interface{}) (*http.Request, error)
	Do(ctx context.Context, request *http.Request, v interface{}, errFn func(int) error) (*http.Response, error)
}

// DefaultBackend returns an instance of the default Backend configuration.
func DefaultBackend() Backend {
	return &defaultBackend{client: &http.Client{}, base: baseURL}
}

type defaultBackend struct {
	client *http.Client
	base   string
}

func (b *defaultBackend) NewRequest(method, path string, query url.Values, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	url := b.base
	if path[0] != '/' {
		url += "/"
	}
	url += path
	if q := query.Encode(); q != "" {
		url += "?" + q
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (b *defaultBackend) Do(ctx context.Context, request *http.Request, v interface{}, errFn func(int) error) (*http.Response, error) {
	request = request.WithContext(ctx)

	resp, err := b.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		if errFn == nil {
			return nil, nil
		}

		apiErr := errFn(resp.StatusCode)
		if apiErr == nil {
			return nil, nil
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(apiErr); err != nil {
			return nil, err
		}
		return nil, apiErr
	}

	if v != nil {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(v); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

type endpoint struct {
	backend Backend
}

// Client is an API client for all endpoints.
type Client struct {
	common endpoint // Reuse a single struct instead of allocating one for each endpoint on the heap.

	Pets *PetsClient
}

// New returns a new Client with the default configuration.
func New() *Client {
	c := &Client{}
	c.common.backend = DefaultBackend()

	c.Pets = (*PetsClient)(&c.common)

	return c
}