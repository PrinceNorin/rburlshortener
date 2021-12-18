package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PrinceNorin/rburlshortener/service"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Create(input service.ShortURLInput) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

func (m *mockService) FindURLs(params *service.FindParams) (*service.Result, error) {
	args := m.Called(params)
	return args.Get(0).(*service.Result), args.Error(1)
}

func (m *mockService) Delete(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *mockService) IncreaseHitCount(code string) error {
	return nil
}

func (m *mockService) GetFullURL(code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}

func TestCreateShortURLHandler(t *testing.T) {
	type testRequest struct {
		url       string
		expiresIn int64
	}
	type testResponse struct {
		code   string
		status int
		err    error
		body   string
	}
	type test struct {
		req testRequest
		res testResponse
	}

	tests := []test{
		{
			req: testRequest{url: "http://example.com"},
			res: testResponse{
				code:   "123",
				status: 201,
				err:    nil,
				body:   `{"url":"http://127.0.0.1/123"}`,
			},
		},
		{
			req: testRequest{url: "http://example.com", expiresIn: 3600},
			res: testResponse{
				code:   "456",
				status: 201,
				err:    nil,
				body:   `{"url":"http://127.0.0.1/456"}`,
			},
		},
		{
			req: testRequest{url: "example.com"},
			res: testResponse{
				code:   "",
				status: 400,
				err:    service.ErrInvalidURL,
				body:   `{"error":["invalid url"]}`,
			},
		},
		{
			req: testRequest{url: "http://example.com", expiresIn: -1},
			res: testResponse{
				code:   "",
				status: 400,
				err:    service.ErrInvalidExpiresIn,
				body:   `{"error":["invalid expires in"]}`,
			},
		},
	}

	mockSvc := new(mockService)
	h := NewHTTPHandler(HTTPConfig{
		ServerHost: "http://127.0.0.1",
		Service:    mockSvc,
	})

	for _, tc := range tests {
		body := fmt.Sprintf(`{"url": "%s", "expiresIn": %d}`, tc.req.url, tc.req.expiresIn)
		req, err := http.NewRequest("POST", "/", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		mockSvc.On("Create", service.ShortURLInput{
			URL:       tc.req.url,
			ExpiresIn: tc.req.expiresIn,
		}).Return(tc.res.code, tc.res.err)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if status := rr.Code; status != tc.res.status {
			t.Errorf("handler returned wrong status code: expected %v, got %v",
				tc.res.status, status)
		}
		if body := strings.TrimSpace(rr.Body.String()); body != tc.res.body {
			t.Errorf("handler returned wrong response: expected %v, got %v",
				tc.res.body, body)
		}

		mockSvc.AssertExpectations(t)
	}
}

func TestGetFullURLHandler(t *testing.T) {
	mockSvc := new(mockService)
	h := NewHTTPHandler(HTTPConfig{
		ServerHost: "http://127.0.0.1",
		Service:    mockSvc,
	})

	type test struct {
		input string
		want  int
	}

	tests := []test{
		{input: "123", want: 302},
		{input: "456", want: 410},
		{input: "789", want: 404},
	}

	mockSvc.On("GetFullURL", "123").Return("http://example.com", nil)
	mockSvc.On("GetFullURL", "456").Return("", service.ErrShortURLExpired)
	mockSvc.On("GetFullURL", "789").Return("", service.ErrRecordNotFound)

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/"+tc.input, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRecorder()
		h.ServeHTTP(r, req)
		if status := r.Code; status != tc.want {
			t.Errorf("handler returned wrong status code: expected %v, got %v", tc.want, status)
		}
	}

	mockSvc.AssertExpectations(t)
}

func TestAdminListShortURLsHandler(t *testing.T) {
	mockSvc := new(mockService)
	h := NewHTTPHandler(HTTPConfig{
		ServerHost: "http://127.0.0.1",
		Service:    mockSvc,
		AdminToken: "1234",
	})

	type response struct {
		Data       []*service.ShortURL `json:"data"`
		TotalCount int64               `json:"totalCount"`
	}
	type test struct {
		params map[string]string
		token  string
		resp   interface{}
		status int
	}

	expiresAt := time.Now().UTC()
	shortURLs := []*service.ShortURL{
		{
			FullURL: "http://example.com",
			Domain:  "example.com",
			Code:    "123",
		},
		{
			FullURL:   "http://example1.com",
			Domain:    "example1.com",
			Code:      "456",
			ExpiresAt: &expiresAt,
		},
	}

	tests := []test{
		{
			token: "1234",
			params: map[string]string{
				"size":   "10",
				"offset": "0",
			},
			resp: &response{
				Data:       shortURLs,
				TotalCount: 2,
			},
			status: 200,
		},
		{
			token: "1234",
			params: map[string]string{
				"size":      "10",
				"offset":    "0",
				"shortCode": "123",
			},
			resp: &response{
				Data:       []*service.ShortURL{shortURLs[0]},
				TotalCount: 1,
			},
			status: 200,
		},
		{
			token:  "invalid token",
			status: 403,
			resp:   map[string]string{"error": "403 Forbidden!"},
		},
	}

	mockSvc.On("FindURLs", &service.FindParams{
		Offset: 0,
		Size:   10,
		Filter: &service.FilterParams{},
	}).Return(&service.Result{
		Data:       shortURLs,
		TotalCount: 2,
	}, nil)
	mockSvc.On("FindURLs", &service.FindParams{
		Offset: 0,
		Size:   10,
		Filter: &service.FilterParams{
			Code: "123",
		},
	}).Return(&service.Result{
		Data:       []*service.ShortURL{shortURLs[0]},
		TotalCount: 1,
	}, nil)

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/admin/shortUrls", nil)
		if err != nil {
			t.Fatal(err)
		}

		q := req.URL.Query()
		for key, val := range tc.params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
		req.Header.Add("Authorization", "Bearer "+tc.token)

		r := httptest.NewRecorder()
		h.ServeHTTP(r, req)

		if status := r.Code; status != tc.status {
			t.Errorf("handler returned wrong status code: expected %v, got %v", tc.status, status)
		}

		buf, err := json.Marshal(tc.resp)
		if err != nil {
			t.Fatal(err)
		}
		expectedResp := strings.TrimSpace(string(buf))
		actualResp := strings.TrimSpace(r.Body.String())
		if actualResp != expectedResp {
			t.Errorf("handler returned wrong response: expected %v, got %v", expectedResp, actualResp)
		}
	}

	mockSvc.AssertExpectations(t)
}

func TestAdminDeleteShortURL(t *testing.T) {
	mockSvc := new(mockService)
	h := NewHTTPHandler(HTTPConfig{
		ServerHost: "http://127.0.0.1",
		Service:    mockSvc,
		AdminToken: "1234",
	})

	mockSvc.On("Delete", "123").Return(nil)
	mockSvc.On("Delete", "456").Return(service.ErrRecordNotFound)

	type test struct {
		code   string
		token  string
		status int
	}

	tests := []test{
		{code: "123", status: 204, token: "1234"},
		{code: "456", status: 404, token: "1234"},
		{code: "123", status: 403, token: "invalid"},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("DELETE", "/admin/shortUrls/"+tc.code, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer "+tc.token)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, req)

		if status := r.Code; status != tc.status {
			t.Errorf("handler returned wrong status code: expected %v, got %v", tc.status, status)
		}
	}

	mockSvc.AssertExpectations(t)
}
