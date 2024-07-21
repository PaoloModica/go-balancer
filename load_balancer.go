package load_balancer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type RequestResult struct {
	StatusCode int
	Message    []byte
}

type LoadBalancerServer struct {
	http.Handler
}

func NewLoadBalancerServer(addr string, port int) *LoadBalancerServer {
	resultsChannel := make(chan RequestResult)
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(forwardToBackend(addr, port, resultsChannel)))

	return &LoadBalancerServer{router}
}

func forwardToBackend(address string, port int, results chan RequestResult) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		go func(r *http.Request) {
			url := fmt.Sprintf("http://%s:%d%s", address, port, r.URL.Path)
			req, err := http.NewRequest(r.Method, url, r.Body)
			if err != nil {
				log.Printf("an error occured while packing request to back-end server: %s", err)
				results <- RequestResult{StatusCode: http.StatusInternalServerError, Message: []byte("internal server error")}
				return
			}
			client := http.Client{Timeout: 30 * time.Second}
			defer client.CloseIdleConnections()

			res, err := client.Do(req)

			if err != nil {
				log.Printf("an error occured while forwarding request to back-end server: %s", err)
				results <- RequestResult{StatusCode: http.StatusInternalServerError, Message: []byte("internal server error")}
				return
			}

			responseBody, err := io.ReadAll(res.Body)

			if err != nil {
				log.Printf("an error occurred while parsing response: %s", err)
				results <- RequestResult{StatusCode: http.StatusInternalServerError, Message: []byte("internal server error")}
				return
			}

			results <- RequestResult{StatusCode: res.StatusCode, Message: responseBody}
		}(r)

		backendRes := <-results
		w.WriteHeader(backendRes.StatusCode)
		w.Write(backendRes.Message)
	}
}
