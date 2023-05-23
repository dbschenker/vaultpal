package testutil

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"net/http"
	"net/http/httptest"
	"testing"
)

type VaultServerMock struct {
	T                   *testing.T
	Server              *httptest.Server
	CloseServer         func()
	ServeLookupSelfFunc func()
	ServeMocks          map[string]ServeMockFunc
}

type ServeMockFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)

func NewVaultServerMock(t *testing.T) *VaultServerMock {
	o := new(VaultServerMock)
	o.T = t
	o.Server = httptest.NewServer(o)
	o.CloseServer = o.Server.Close
	o.ServeMocks = map[string]ServeMockFunc{}

	return o
}

func (o *VaultServerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	serveMockFunc, ok := o.ServeMocks[r.URL.Path]
	if ok {
		serveMockFunc(o.T, w, r)
		return
	} else {
		o.T.Fatalf("unexpected path: %q", r.URL.Path)
	}
}

func WriteJsonResponse(t *testing.T, jsonO interface{}, w http.ResponseWriter) {
	json, err := json.Marshal(jsonO)
	if err != nil {
		t.Fatal(err)
	}
	w.Write(json)
}

type MockErrorData struct {
	ErrorBody  string
	Errors     *[]string
	HTTPStatus int
}

func (m *MockErrorData) MockErrorResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := api.ErrorResponse{Errors: *m.Errors}
	w.WriteHeader(m.HTTPStatus)
	WriteJsonResponse(t, resp, w)
}
