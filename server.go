package load_balancer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HealthcheckStatus struct {
	Status string
}

type BasicServer struct {
	http.Handler
}

func NewServer() *BasicServer {
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(homeRouteHandler))
	router.Handle("/healthcheck", http.HandlerFunc(healthcheckRouteHandler))

	server := BasicServer{router}

	return &server
}

func homeRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request from %v\n", r.Header.Get("Host"))
	fmt.Printf("%v %v %v\n", r.Method, r.URL, r.Proto)
	fmt.Printf("Host %v:\n", r.Host)
	fmt.Printf("User-Agent: %v\n", r.UserAgent())
	fmt.Printf("Accept: %v\n", r.Header.Get("Accept-Encoding"))

	fmt.Fprintf(w, "hello %v", r.Host)
}

func healthcheckRouteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	status := HealthcheckStatus{"READY"}
	json.NewEncoder(w).Encode(status)
}
