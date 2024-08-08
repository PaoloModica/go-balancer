package load_balancer_test

import (
	"encoding/json"
	load_balancer "go-balancer"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func assertResponseStatusCode(t *testing.T, expected int, got int) {
	t.Helper()

	if expected != got {
		t.Errorf("expected %d, got %d", expected, got)
	}
}

func assertResponseBody(t *testing.T, expected, got load_balancer.HealthcheckStatus) {
	t.Helper()

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected response body %v, got %v", expected, got)
	}
}

func TestServer(t *testing.T) {
	t.Run("GET / returns 200 OK", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		server := load_balancer.NewServer()

		server.ServeHTTP(response, request)

		expectedResult := 200
		gotStatusCode := response.Result().StatusCode

		assertResponseStatusCode(t, expectedResult, gotStatusCode)
	})
	t.Run("GET /healthcheck returns 200 OK and status information", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/healthcheck", nil)
		response := httptest.NewRecorder()
		server := load_balancer.NewServer()

		server.ServeHTTP(response, request)

		expectedStatusCode := 200
		expectedResponseBody := load_balancer.HealthcheckStatus{"READY"}

		responseResult := response.Result()
		gotStatusCode := responseResult.StatusCode
		var gotResponseBody load_balancer.HealthcheckStatus

		err := json.NewDecoder(response.Body).Decode(&gotResponseBody)

		if err != nil {
			t.Fatal("an error occurred while decoding response body")
		}

		assertResponseStatusCode(t, expectedStatusCode, gotStatusCode)

		assertResponseBody(t, expectedResponseBody, gotResponseBody)
	})
}
