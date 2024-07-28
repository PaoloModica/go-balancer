package main

import (
	load_balancer "go-balancer"
	"net/http"
	"os"
)

func main() {
	serverAddresses := []string{}
	serverAddresses = append(serverAddresses, os.Args[1:]...)

	loadBalancerServer := load_balancer.NewLoadBalancerServer(serverAddresses)
	http.ListenAndServe(":5000", loadBalancerServer)
}
