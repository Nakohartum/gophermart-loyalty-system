package main

import (
	"fmt"
	"net/http"
)

func main() {
	parseFlags()
	mux := http.NewServeMux()
	fmt.Println("Server started")
	if err := http.ListenAndServe(ConfigData.RunAddress, mux); err != nil {
		fmt.Println(err)
	}
	
}
