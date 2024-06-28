package load_balancer

import (
	"fmt"
	"net/http"
	"os"
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
	fmt.Printf("Received request from %v", r.Header.Get("Host"))
	r.WriteProxy(os.Stdout)
}
