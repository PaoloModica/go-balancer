package main

import (
	load_balancer "go-balancer"
	"net/http"
)

func main() {
	loadBalancerServer := load_balancer.NewLoadBalancerServer("127.0.0.1", 5001)
	http.ListenAndServe(":5000", loadBalancerServer)
}
