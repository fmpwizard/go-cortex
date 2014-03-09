package main

import (
	"flag"
	"fmt"
	"github.com/fmpwizard/go-cortex/services"
	"log"
	"net/http"
)

var httpPort string

//You need to get a wit access token to use their services
func init() {
	flag.StringVar(&httpPort, "httpPort", "7070", "Port number to listen for questions.")
}

func main() {
	flag.Parse()
	listenForCommands()

	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}

//listenForCommands starts listening for the Arduino command
// to start recording a voice command.
//Once we get a voice command, the ARduino service will
//send it to Wit and then the intent channel will get the
// go struct with the information we need.
//We then call ProcessIntent and start listening for
//a new Arduino command.
func listenForCommands() {
	log.Println("1- listening for commands.")
	intent := make(chan services.WitMessage)
	go services.ArduinoIn(intent)
	go func() {
		select {
		case ret := <-intent:
			log.Printf("Got intent %+v\n", ret)
			ProcessIntent(ret)
			listenForCommands()
		}
	}()
}

func handler(w http.ResponseWriter, r *http.Request) {
	//read the "q" GET query parameter and pass it to
	// the wit service
	message := r.FormValue("q")
	if len(message) > 0 {
		ret := ProcessIntent(services.FetchIntent(message))
		//print what we understood from your request to the browser.
		fmt.Fprintf(w, ret)
	} else {
		fmt.Fprintf(w, "Please add a ?q=<text here> to the url")
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
		services.Arduino(action, light)
		return fmt.Sprintf("Turning light %v %s", light, action)
	}
	return ""
}
