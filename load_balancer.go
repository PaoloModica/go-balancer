package load_balancer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type RequestResult struct {
	StatusCode int
	Message    []byte
}

type BackendServerScheduleManager struct {
	ServerList        []string
	NextServerId      int
	NextServerIdMutex *sync.Mutex
}

type LoadBalancerServer struct {
	http.Handler
}

func (bm *BackendServerScheduleManager) GetNextServerAddress() string {
	bm.NextServerIdMutex.Lock()
	defer bm.NextServerIdMutex.Unlock()

	var i int

	if bm.NextServerId <= len(bm.ServerList)-1 {
		i = bm.NextServerId
		bm.NextServerId = bm.NextServerId + 1
	} else {
		i = 0
		bm.NextServerId = 0
	}

	baseUrl := fmt.Sprintf("http://%s", bm.ServerList[i])

	return baseUrl
}

func NewLoadBalancerServer(addresses []string) *LoadBalancerServer {
	resultsChannel := make(chan RequestResult)

	bm := BackendServerScheduleManager{
		addresses,
		0,
		&sync.Mutex{},
	}

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(forwardToBackend(bm, resultsChannel)))
	log.Printf("back-end server addresses: %v", addresses)
	return &LoadBalancerServer{router}
}

func forwardToBackend(bm BackendServerScheduleManager, results chan RequestResult) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		go func(r *http.Request) {
			baseUrl := bm.GetNextServerAddress()
			log.Printf("next back-end service address, %s", baseUrl)
			url := fmt.Sprintf("%s%s", baseUrl, r.URL.Path)
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
