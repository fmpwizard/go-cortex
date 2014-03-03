package main

import (
	"flag"
	"github.com/fmpwizard/doodit/services"
	"log"
)

var message string

func init() {
	flag.StringVar(&message, "q", "Turn the lights on", "Message to process")
}

func main() {
	flag.Parse()
	ProcessIntent(services.FetchIntent(message))
}

func ProcessIntent(jsonResponse services.WitMessage) {
	switch jsonResponse.Outcome.Intent {
	case "lights":
		light := jsonResponse.Outcome.Entities.Number.Value
		action := jsonResponse.Outcome.Entities.OnOff.Value
		log.Printf("Turning light %v %s", light, action)
		services.Arduino(action, light)
	}
}
