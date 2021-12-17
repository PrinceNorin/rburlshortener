package transport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	return nil, nil
}

func (m *mockService) Delete(code string) error {
	return nil
}

func (m *mockService) IncreaseHitCount(code string) error {
	return nil
}

func (m *mockService) GetFullURL(code string) (string, error) {
	return "", nil
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
