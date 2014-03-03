package main

import (
	"flag"
	"fmt"
	"github.com/fmpwizard/doodit/services"
	"net/http"
)

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	//read the "q" GET query parameter and pass it to
	// the wit service
	message := r.FormValue("q")
	if len(message) > 0 {
		ret := ProcessIntent(services.FetchIntent(message))
		//print what we understood from your request to the browser.
		fmt.Fprintf(w, ret)
	}
}

//ProcessIntent gets the json parsed result from wit.ai and
//depending on the intent, it calles the right service.
//So far we only have one service, the Arduino lights service
func ProcessIntent(jsonResponse services.WitMessage) string {
	switch jsonResponse.Outcome.Intent {
	case "lights":
		light := jsonResponse.Outcome.Entities.Number.Value
		action := jsonResponse.Outcome.Entities.OnOff.Value
		ret := fmt.Sprintf("Turning light %v %s", light, action)
		services.Arduino(action, light)
		return ret
	}
	return ""
}
