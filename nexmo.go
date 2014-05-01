package main

import (
	"log"
	"net/http"
)

//NexmoHandler handles the GET requests from nexmo
func NexmoHandler(w http.ResponseWriter, r *http.Request) {
	//A sample request from the nexmo service is
	//?msisdn=19150000001&to=12108054321
	//&messageID=000000FFFB0356D1&text=This+is+an+inbound+message
	//&type=text&message-timestamp=2012-08-19+20%3A38%3A23
	//So we read all those parameters
	messageID := r.FormValue("messageID")
	text := r.FormValue("text")
	typ := r.FormValue("type")
	timestamp := r.FormValue("message-timestamp=")
	if len(text) > 0 && typ == "text" {
		intent, err := FetchIntent(text)
		if err != nil {
			log.Printf("Error: %+v", err)
		} else {
			ret := ProcessIntent(intent)
			log.Printf("We got messageID: %v on %v ", messageID, timestamp)
			log.Printf("Wit gave us: %+v ", ret)
		}

	} else {
		log.Print("Error: we got a blank text message")
	}
	w.WriteHeader(http.StatusOK)
}
