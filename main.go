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

	go func() {
		ListenStream()
	}()

	http.HandleFunc("/wit", WitHandler)
	http.HandleFunc("/sms", NexmoHandler)
	http.ListenAndServe(fmt.Sprintf(":%v", config.HttpPort), nil)
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

type CortexConfig struct {
	HttpPort            string
	FlowdockAccessToken string
	WitAccessToken      string
	Flows               string
	FlowsTicketsUrls    []map[string]string
}
