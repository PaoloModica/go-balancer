package main

import (
	"flag"
	load_balancer "go-balancer"
	"net/http"
	"time"
)

func main() {
	var healthcheckPeriod int
	var healthcheckRoute string

	flag.IntVar(&healthcheckPeriod, "htime", 30, "healthcheck period in seconds")
	flag.StringVar(&healthcheckRoute, "hroute", "/healthcheck", "healthcheck route")
	flag.Parse()

	serverAddresses := []string{}
	serverAddresses = append(serverAddresses, flag.Args()...)

	serverHealthChecker := load_balancer.BackendHealthCheckerFunc(load_balancer.HealthChecker)
	loadBalancerServer := load_balancer.NewLoadBalancerServer(serverAddresses, serverHealthChecker, time.Duration(healthcheckPeriod)*time.Second, healthcheckRoute)
	http.ListenAndServe(":5000", loadBalancerServer)
}
