package main

import (
	load_balancer "go-balancer"
	"net/http"
)

func main() {
	backendServer := load_balancer.NewServer()
	http.ListenAndServe(":5001", backendServer)
}
