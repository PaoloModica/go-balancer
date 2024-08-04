package load_balancer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type BackendHealthChecker interface {
	ScheduleHealthCheck(duration time.Duration, route string, bm *BackendServerScheduleManager)
}

type BackendHealthCheckerFunc func(time.Duration, string, *BackendServerScheduleManager)

func (b BackendHealthCheckerFunc) ScheduleHealthCheck(duration time.Duration, route string, bm *BackendServerScheduleManager) {
	b(duration, route, bm)
}

func HealthChecker(duration time.Duration, route string, bm *BackendServerScheduleManager) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for t := range ticker.C {
		log.Printf("Health checking at %v", t)
		bm.CheckServerHealth(route, 30*time.Second)
	}
}

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

	if bm.NextServerId > len(bm.ServerList)-1 {
		bm.NextServerId = 0
	}

	baseUrl := fmt.Sprintf("http://%s", bm.ServerList[bm.NextServerId])
	bm.NextServerId += 1

	return baseUrl
}

func (bm *BackendServerScheduleManager) CheckServerHealth(healthcheckRoute string, timeout time.Duration) {
	results := make(chan RequestResult, len(bm.ServerList))
	var wg sync.WaitGroup

	log.Printf("health checking back-end: %v", bm.ServerList)

	for i, baseUrl := range bm.ServerList {
		wg.Add(1)
		go func(baseUrl string, i int) {
			defer wg.Done()

			url := fmt.Sprintf("http://%s%s", baseUrl, healthcheckRoute)
			log.Printf("sending request to %s", url)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.Printf("an error occured while sending %s request to back-end server: %s", url, err)
				results <- RequestResult{StatusCode: http.StatusInternalServerError, Message: []byte(fmt.Sprint(i))}
				return
			}
			client := http.Client{Timeout: timeout}
			defer client.CloseIdleConnections()

			res, err := client.Do(req)

			if err != nil {
				log.Printf("an error occured while forwarding request to %s: %s", url, err)
				results <- RequestResult{StatusCode: http.StatusInternalServerError, Message: []byte(fmt.Sprint(i))}
				return
			}

			log.Printf("%s status: %d", url, res.StatusCode)

			results <- RequestResult{StatusCode: res.StatusCode, Message: []byte(fmt.Sprint(i))}
		}(baseUrl, i)
	}

	wg.Wait()
	close(results)

	bm.NextServerIdMutex.Lock()
	defer bm.NextServerIdMutex.Unlock()

	for r := range results {
		if r.StatusCode != http.StatusOK {
			index, err := strconv.Atoi(string(r.Message))

			if err != nil {
				log.Printf("an error occured while checking result %s", err)
			}

			bm.ServerList = append(bm.ServerList[:index], bm.ServerList[index+1:]...)
		}
	}
	log.Printf("healthcheck completed, online services: %v", bm.ServerList)
}

func NewLoadBalancerServer(addresses []string, healthchecker BackendHealthChecker, healthcheckPeriod time.Duration, healthcheckRoute string) *LoadBalancerServer {
	resultsChannel := make(chan RequestResult)

	bm := BackendServerScheduleManager{
		addresses,
		0,
		&sync.Mutex{},
	}

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(forwardToBackend(bm, resultsChannel)))
	log.Printf("back-end server addresses: %v", addresses)

	go healthchecker.ScheduleHealthCheck(healthcheckPeriod, healthcheckRoute, &bm)

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
