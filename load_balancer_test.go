package load_balancer_test

import (
	"fmt"
	load_balancer "go-balancer"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type loadBalancerTestCase struct {
	method string
	route  string
}

func testServer(t *testing.T, addr string, port int) *httptest.Server {
	t.Helper()

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))

	if err != nil {
		t.Errorf("an error occurred while creating test server: %v", err)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "hello %v", r.Host)
	})

	ts := httptest.NewUnstartedServer(h)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	return ts
}

func TestLoadBalancer(t *testing.T) {
	testCases := []loadBalancerTestCase{
		{http.MethodGet, "/ "},
		{http.MethodPost, "/"},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s %s forwards the request to back-end service and returns 200 OK", testCase.method, testCase.route), func(t *testing.T) {
			tsAddr := "127.0.0.1"
			tsPort := 5000

			loadBalancer := load_balancer.NewLoadBalancerServer(tsAddr, tsPort)

			ts := testServer(t, tsAddr, tsPort)
			defer ts.Close()

			request, _ := http.NewRequest(testCase.method, testCase.route, nil)
			response := httptest.NewRecorder()

			loadBalancer.ServeHTTP(response, request)

			expectedResult := 200
			gotStatusCode := response.Result().StatusCode

			if expectedResult != gotStatusCode {
				t.Errorf("Expected %d, got %d", expectedResult, gotStatusCode)
			}
		})
	}
}
