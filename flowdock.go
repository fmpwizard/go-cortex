package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var availableFlows []Flows

//FetchFlows fetches all the flows we have access to
func FetchFlows() {
	url := fmt.Sprintf("https://%+v@api.flowdock.com/flows", config.FlowdockAccessToken)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting list of flows: %v", err)
	}
	FlowsParse(res.Body)
}

//FlowsParse parses the json response from flowdock into a struct
func FlowsParse(body io.ReadCloser) []Flows {
	flowsAsJon, _ := ioutil.ReadAll(body)

	if ok := json.Unmarshal(flowsAsJon, &availableFlows); ok != nil {
		log.Printf("Error parsing flows data %+v", ok)
	}
	log.Printf("flow data is %+v", availableFlows)
	return availableFlows
}

//GetFlowURL given a flow id as string, return the url for the flow
func GetFlowURL(id string) (string, error) {
	for _, flow := range availableFlows {
		if flow.Id == id {
			return flow.Url, nil
		}
	}
	return "", errors.New("Flow url not found by key " + id)
}

//GetFlowName given a flow id as string, return the name of the flow
func GetFlowName(id string) (string, error) {
	for _, flow := range availableFlows {
		if flow.Id == id {
			return flow.Parameterized_name, nil
		}
	}
	return "", errors.New("Flow url not found by key " + id)
}

//GetIssueURLForFlowName given a flow name, return the issues url for it
func GetIssueURLForFlowName(parametizedName string) (string, error) {
	for _, row := range config.FlowsTicketsUrls {
		url, ok := row[parametizedName]
		if ok {
			return url, nil
		}
	}
	return "", errors.New("Could not find issue url for flow: " + parametizedName)
}

//ListenStream starts pulling the flowdock stream api
func ListenStream() {
	FetchFlows()
	var flowMessage FlowdockMsg
	var flowUpdatedMessage FlowdockUpdatedMsg
	var flowComment FlowdockComment

	url := fmt.Sprintf("https://stream.flowdock.com/flows?filter=%s", config.Flows)
	token := []byte(config.FlowdockAccessToken)
	str := base64.StdEncoding.EncodeToString(token)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", str))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	if res.StatusCode == 401 {
		log.Fatalln("Got 401 from flowdock.")
	}

	defer res.Body.Close()
	reader := bufio.NewReader(res.Body)
	for {
		line, _ := reader.ReadBytes('\r')
		line = bytes.TrimSpace(line)
		jsonString := string(line[:])
		_ = jsonString
		if len(jsonString) < 4 {
			log.Fatalln("Got empty response from glowdock, shutting down")
			os.Exit(1)
		}
		//log.Printf("Flowdock stream string response: %v\n\n", jsonString)
		json.Unmarshal(line, &flowMessage)
		var parentMessageID = flowMessage.Id

		switch flowMessage.Event {
		case "message-edit":
			json.Unmarshal(line, &flowUpdatedMessage)
			//log.Printf("parsed1: %+v\n\n", flowUpdatedMessage)
			ret := ProcessIntent(FetchIntent(flowUpdatedMessage.Content.Updated_content))
			replyToFlow(ret, flowUpdatedMessage.Id, flowUpdatedMessage.Flow)
		case "message":
			//log.Printf("parsed2: %+v\n\n", flowMessage)
			ret := ProcessIntent(FetchIntent(flowMessage.Content))
			replyToFlow(ret, flowMessage.Id, flowMessage.Flow)
		case "comment":
			if flowMessage.User != "77156" {
				json.Unmarshal(line, &flowComment)
				//log.Printf("parsed2: %+v\n\n", flowComment)
				ret := ProcessIntent(FetchIntent(flowComment.Content.Text))
				for _, v := range flowComment.Tags {
					if strings.Contains(v, "influx") {
						parentID, _ := strconv.ParseInt(strings.Split(v, ":")[1], 0, 64)
						parentMessageID = parentID
					}
				}
				replyToFlow(ret, parentMessageID, flowComment.Flow)
			} else {
				//log.Println("skipping Cortex's message.")
			}
		}
	}
}

func replyToFlow(ret WitResponse, originalMessageID int64, flowID string) {
	if ret.Temperature.Unit != "" {
		handleTemperature(ret, originalMessageID, flowID)
	} else if len(ret.Github.issues) > 0 {
		handleGithub(ret, originalMessageID, flowID)
	}

}

func handleTemperature(ret WitResponse, originalMessageID int64, flowID string) {
	temperature := ret.Temperature.Degrees
	switch ret.Temperature.Unit {
	case "C":
		flowdockPost(fmt.Sprintf("Which is %+vF", cToF(temperature)), originalMessageID, flowID)
	case "F":
		flowdockPost(fmt.Sprintf("Which is %+vC", fToC(temperature)), originalMessageID, flowID)
	}
}

func handleGithub(ret WitResponse, originalMessageID int64, flowID string) {
	for _, issue := range ret.Github.issues {
		flowParametizedName, error := GetFlowName(flowID)
		if error != nil {
			log.Printf("Error trying to get parametized flow name for id %v", flowID)
		}
		issueURL, error := GetIssueURLForFlowName(flowParametizedName)
		if error != nil {
			log.Printf("%s", error)
		}

		//log.Println("\n\n\n\n\n\n\n\nflowParametizedName ", flowParametizedName)
		//log.Println("issueURL ", issueURL)
		url := fmt.Sprintf("just click here: %+v%+v", issueURL, issue)
		flowdockPost(url, originalMessageID, flowID)
	}
}

func flowdockPost(message string, originalMessageID int64, flowID string) {
	flowURL, err := GetFlowURL(flowID)
	if err != nil {
		log.Panicf("Error getting flow id, got %v", err)
	}
	url := fmt.Sprintf("%+v/messages/%+v/comments", flowURL, originalMessageID)
	token := []byte(config.FlowdockAccessToken)
	str := base64.StdEncoding.EncodeToString(token)
	client := &http.Client{}
	payload := []byte(`{
	  "event": "comment",
	  "content":"` + message + `"}`)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", str))
	req.Header.Add("Content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Posting a message to Flowdock gave: %v", err)
	}

	defer res.Body.Close()
	//log.Printf("sending %+v", message)
	value, err := ioutil.ReadAll(res.Body)
	_ = value
	if err != nil {
		log.Fatalf("Could not read body, got: %v", err)
	}

	//aa := string(value[:])
	//st := string(payload[:])
	//log.Printf("sending %+v  %+v %+v", st, err, aa)

	if res.StatusCode == 401 {
		log.Fatalln("Access denied, check your wit access token ")
	}
}

func fToC(f int) int {
	return (f - 32) * 5 / 9
}

func cToF(c int) int {
	return c*9/5 + 32
}

//FlowdockMsg struct all the information we care about from flowdock message of type message
type FlowdockMsg struct {
	Event   string
	Tags    []string
	Uuid    string
	Persist bool
	Id      int64
	Flow    string
	Content string
	Sent    int64
	User    string
}

//FlowdockUpdatedMsg struct all the information we care about from flowdock message of type update message
type FlowdockUpdatedMsg struct {
	Event   string
	Tags    []string
	Uuid    string
	Persist bool
	Id      int64
	Flow    string
	Content flowdockContent
	Sent    int64
	User    string
}

type flowdockContent struct {
	Message         float64
	Updated_content string
}

//FlowdockComment struct all the information we care about from flowdock message of type comment
type FlowdockComment struct {
	Event   string
	Tags    []string
	Uuid    string
	Persist bool
	Id      int64
	Flow    string
	Content flowdockCommentContent
	Sent    int64
	User    string
}

type flowdockCommentContent struct {
	Title string
	Text  string
}

//Flows struct that holds information about all the flows we can access
type Flows struct {
	Id                 string
	Name               string
	Parameterized_name string
	Email              string
	Description        string
	Url                string
	Web_url            string
	Unread_mentions    int
}
