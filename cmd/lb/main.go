package main

import (
	load_balancer "go-balancer"
	"net/http"
)

func main() {
	loadBalancerServer := load_balancer.NewServer()
	http.ListenAndServe(":5000", loadBalancerServer)
}
