package load_balancer_test

import (
	"fmt"
	load_balancer "go-balancer"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type loadBalancerTestCase struct {
	method string
	route  string
}

func testServer(t *testing.T, addr string) *httptest.Server {
	t.Helper()

	l, err := net.Listen("tcp", addr)

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

type SpyBackendHealthChecker struct {
	calls int
}

func (s *SpyBackendHealthChecker) ScheduleHealthCheck(duration time.Duration, route string, bm *load_balancer.BackendServerScheduleManager) {
	s.calls += 1
}

func TestBackendServerScheduleManager(t *testing.T) {
	serverList := []string{
		"127.0.0.1:5001",
		"127.0.0.1:5002",
	}
	mu := &sync.Mutex{}
	t.Run("get next server address, backendServerId less than length of list of address", func(t *testing.T) {
		bm := load_balancer.BackendServerScheduleManager{
			serverList,
			0,
			mu,
		}
		expectedBaseUrl := "http://127.0.0.1:5001"
		gotBaseUrl := bm.GetNextServerAddress()

		if expectedBaseUrl != gotBaseUrl {
			t.Errorf("expected %s, got %s base URL string", expectedBaseUrl, gotBaseUrl)
		}
	})
	t.Run("get next server address, backendServerId equal to length of list of address", func(t *testing.T) {
		bm := load_balancer.BackendServerScheduleManager{
			serverList,
			2,
			mu,
		}
		expectedBaseUrl := "http://127.0.0.1:5001"
		gotBaseUrl := bm.GetNextServerAddress()

		if expectedBaseUrl != gotBaseUrl {
			t.Errorf("expected %s, got %s base URL string", expectedBaseUrl, gotBaseUrl)
		}
	})
	t.Run("check server status, server 1 healthy, server 2 out", func(t *testing.T) {
		bm := load_balancer.BackendServerScheduleManager{
			serverList,
			2,
			mu,
		}
		ts := testServer(t, serverList[0])
		defer ts.Close()

		healthcheckRoute := "/healthcheck"
		timeout := 3 * time.Second
		bm.CheckServerHealth(healthcheckRoute, timeout)

		expectedServerListLength := 1

		if len(bm.ServerList) != expectedServerListLength {
			t.Errorf("expected server list length equal to %d, found %d", expectedServerListLength, len(bm.ServerList))
		}
	})
}

func TestLoadBalancer(t *testing.T) {
	testCases := []loadBalancerTestCase{
		{http.MethodGet, "/"},
		{http.MethodPost, "/"},
	}
	healthcheckPeriod := 10 * time.Second
	healthcheckRoute := "/healthcheck"

	dummyHealthChecker := &SpyBackendHealthChecker{}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s %s forwards the request to back-end service and returns 200 OK", testCase.method, testCase.route), func(t *testing.T) {
			addresses := []string{
				"127.0.0.1:5001",
				"127.0.0.1:5002",
			}

			loadBalancer := load_balancer.NewLoadBalancerServer(addresses, dummyHealthChecker, healthcheckPeriod, healthcheckRoute)

			ts := testServer(t, addresses[0])
			defer ts.Close()

			request, _ := http.NewRequest(testCase.method, testCase.route, nil)
			response := httptest.NewRecorder()

			loadBalancer.ServeHTTP(response, request)

			expectedResult := 200
			gotStatusCode := response.Result().StatusCode

			if expectedResult != gotStatusCode {
				t.Errorf("expected %d, got %d", expectedResult, gotStatusCode)
			}

			if dummyHealthChecker.calls <= 0 {
				t.Errorf("expected health checking function to have been called")
			}
		})
	}
}
