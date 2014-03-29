package main

import (
	"flag"
	"fmt"
	"github.com/fmpwizard/go-cortex/services"
	"net/http"
)

var httpPort string

//You need to get a wit access token to use their services
func init() {
	flag.StringVar(&httpPort, "httpPort", "7070", "Port number to listen for questions.")
}

func main() {
	flag.Parse()
	http.HandleFunc("/wit", handler)
	http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	//read the "q" GET query parameter and pass it to
	// the wit service
	message := r.FormValue("q")
	if len(message) > 0 {
		ret := services.ProcessIntent(services.FetchIntent(message))
		//print what we understood from your request to the browser.
		fmt.Fprintf(w, ret)
	} else {
		fmt.Fprintf(w, "Please add a ?q=<text here> to the url")
	}
}
