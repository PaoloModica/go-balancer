package load_balancer_test

import (
	load_balancer "go-balancer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	t.Run("GET / returns 200 OK", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		server := load_balancer.NewServer()

		server.ServeHTTP(response, request)

		expectedResult := 200
		gotStatusCode := response.Result().StatusCode

		if expectedResult != gotStatusCode {
			t.Errorf("Expected %d, got %d", expectedResult, gotStatusCode)
		}
	})
}
