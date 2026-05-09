package main

import (
"fmt"
"net/http"
)

func main() {
	mux := http.NewServeMux()
	var server http.Server
	server.Handler = mux
	server.Addr = ":8080"


	fmt.Println("Starting Server @8080")
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
