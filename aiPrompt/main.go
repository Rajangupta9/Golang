package main

import (
	"fmt"
	"log"
	"net/http"
)

type OllamaRequest struct{
	Model string
	Prompt string
	Stream bool
}

func main(){

	
    http.HandleFunc("/prompt", handle)


	var port string = "4000"

    fmt.Printf("server started from %s ", port)
	log.Fatal(http.ListenAndServe(":"+ port, nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	
}