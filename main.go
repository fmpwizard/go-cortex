//Go-Cortex is a robot that uses the wit.ai api to let you control arduinos, as well as interact with flowdock
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var configFile string
var config CortexConfig

func init() {
	flag.StringVar(&configFile, "config", "", "path to cortex.config.json file.")
}

func main() {
	flag.Parse()
	readCortexConfig()

	if config.FlowdockAccessToken != "" {
		go func() {
			listenStream()
		}()

	}
	if config.HttpPort != "" {
		http.HandleFunc("/wit", WitHandler)
		http.HandleFunc("/sms", NexmoHandler)
		http.ListenAndServe(fmt.Sprintf(":%v", config.HttpPort), nil)
	}

}

func readCortexConfig() {
	configBytes, error := ioutil.ReadFile(configFile)
	if error != nil {
		log.Fatalf("Could not read config file, error: %+v", error)
	}
	error = json.Unmarshal(configBytes, &config)
	if error != nil {
		log.Fatalf("Could not parse json file, got: %+v", error)
	}
	log.Printf("Using configuration: %+v", config)
}

//CortexConfig hold the configuration for Cortex to work.
type CortexConfig struct {
	HttpPort            string
	CortexEmail         string
	FlowdockAccessToken string
	WitAccessToken      string
	Flows               string
	FlowsTicketsUrls    []map[string]string
}
