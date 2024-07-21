package load_balancer

import (
	"fmt"
	"net/http"
)

type BasicServer struct {
	http.Handler
}

func NewServer() *BasicServer {
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(homeRouteHandler))

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
