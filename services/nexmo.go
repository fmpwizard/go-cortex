package services

import (
	"log"
	"net/http"
)

//Add a handler for the /sms path
func init() {
	http.HandleFunc("/sms", handler)
}

//Handle the GET requests from nexmo
func handler(w http.ResponseWriter, r *http.Request) {
	//A sample request from the nexmo service is
	//?msisdn=19150000001&to=12108054321
	//&messageId=000000FFFB0356D1&text=This+is+an+inbound+message
	//&type=text&message-timestamp=2012-08-19+20%3A38%3A23
	//So we read all those parameters
	messageId := r.FormValue("messageId")
	text := r.FormValue("text")
	typ := r.FormValue("type")
	timestamp := r.FormValue("message-timestamp=")
	if len(text) > 0 && typ == "text" {
		ret := ProcessIntent(FetchIntent(text))
		log.Printf("We got messageId: %v on %v ", messageId, timestamp)
		log.Printf("Wit gave us: %+v ", ret)
	} else {
		log.Print("Error: we got a blank text message")
	}
	w.WriteHeader(http.StatusOK)
}
