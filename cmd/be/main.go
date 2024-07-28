package main

import (
	"fmt"
	load_balancer "go-balancer"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := 5001
	var err error
	if len(os.Args) > 1 {
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Printf("an error occurred while converting port arg to int: %s", err)
		}
	}
	backendServer := load_balancer.NewServer()
	http.ListenAndServe(fmt.Sprintf(":%d", port), backendServer)
}
