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

var availableFlows []flows

//ListenStream starts pulling the flowdock stream api
func listenStream() {
	fetchFlows()
	res := connectToFlow()
	defer res.Body.Close()
	for {
		processFlowRow(parseFlowRow(bufio.NewReader(res.Body)))
	}
}

//fetchFlows fetches all the flows we have access to
func fetchFlows() {
	url := fmt.Sprintf("https://%+v@api.flowdock.com/flows", config.FlowdockAccessToken)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error getting list of flows: %v", err)
	} else if res.StatusCode != 200 {
		log.Fatalf("got status code %+v", res.StatusCode)
	}

	parseAvailableFlows(res.Body)
	res.Body.Close()
}

func parseAvailableFlows(body io.ReadCloser) {
	flowsAsJon, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalf("error reading body, got: %+v", err)
	}

	if ok := json.Unmarshal(flowsAsJon, &availableFlows); ok != nil {
		log.Fatalf("Error parsing flows data %+v", ok)
	}
}

func connectToFlow() *http.Response {
	url := fmt.Sprintf("https://stream.flowdock.com/flows?filter=%s", config.Flows)
	token := base64.StdEncoding.EncodeToString([]byte(config.FlowdockAccessToken))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln("could not fetch streaming api ", err)
	} else if res.StatusCode != 200 {
		log.Fatalf("got error code: %+v from flowdock.\n", res.StatusCode)
	}

	return res
}

func parseFlowRow(reader *bufio.Reader) (flowdockMsg, []byte) {

	line, err := reader.ReadBytes('\r')
	if err != nil {
		log.Fatalf("something went wrong reading the body: %s", err)
	}
	line = bytes.TrimSpace(line)
	jsonString := string(line[:])
	if len(jsonString) < 4 {
		log.Fatalln("got empty response from flowdock, shutting down")
		os.Exit(1)
	}
	var message flowdockMsg
	json.Unmarshal(line, &message)
	return message, line
}

func processFlowRow(flowMessage flowdockMsg, line []byte) {
	var flowUpdatedMessage flowdockUpdatedMsg
	var flowComment flowdockComment

	switch flowMessage.Event {
	case "message":
		ret := ProcessIntent(FetchIntent(flowMessage.Content))
		replyToFlow(ret, flowMessage.Id, flowMessage.Flow)
	case "message-edit":
		json.Unmarshal(line, &flowUpdatedMessage)
		ret := ProcessIntent(FetchIntent(flowUpdatedMessage.Content.Updated_content))
		replyToFlow(ret, flowUpdatedMessage.Id, flowUpdatedMessage.Flow)
	case "comment":
		if flowMessage.User != "77156" {
			var parentMessageID int64
			json.Unmarshal(line, &flowComment)
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

//getFlowURL given a flow id as string, return the url for the flow
func getFlowURL(id string) (string, error) {
	for _, flow := range availableFlows {
		if flow.Id == id {
			return flow.Url, nil
		}
	}
	return "", errors.New("Flow url not found by key " + id)
}

//getFlowName given a flow id as string, return the name of the flow
func getFlowName(id string) (string, error) {
	for _, flow := range availableFlows {
		if flow.Id == id {
			return flow.Parameterized_name, nil
		}
	}
	return "", errors.New("Flow url not found by key " + id)
}

//getIssueURLForFlowName given a flow name, return the issues url for it
func getIssueURLForFlowName(parametizedName string) (string, error) {
	for _, row := range config.FlowsTicketsUrls {
		url, ok := row[parametizedName]
		if ok {
			return url, nil
		}
	}
	return "", errors.New("Could not find issue url for flow: " + parametizedName)
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
		flowParametizedName, error := getFlowName(flowID)
		if error != nil {
			log.Printf("Error trying to get parametized flow name for id %v", flowID)
		}
		issueURL, error := getIssueURLForFlowName(flowParametizedName)
		if error != nil {
			log.Printf("%s", error)
		}

		url := fmt.Sprintf("just click here: %+v%+v", issueURL, issue)
		flowdockPost(url, originalMessageID, flowID)
	}
}

func flowdockPost(message string, originalMessageID int64, flowID string) {
	flowURL, err := getFlowURL(flowID)
	if err != nil {
		log.Printf("Error getting flow id, got %v", err)
		return
	}
	url := fmt.Sprintf("%+v/messages/%+v/comments", flowURL, originalMessageID)
	token := base64.StdEncoding.EncodeToString([]byte(config.FlowdockAccessToken))
	client := &http.Client{}
	payload := []byte(`{
	  "event": "comment",
	  "content":"` + message + `"}`)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
	req.Header.Add("Content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error posting a message to Flowdock: %v", err)
		return
	} else if res.StatusCode != 200 {
		log.Printf("We got a non 200 code: %+v\n", res.StatusCode)
		return
	}

	defer res.Body.Close()
	value, err := ioutil.ReadAll(res.Body)
	_ = value
	if err != nil {
		log.Fatalf("Could not read body, got: %v", err)
	}
}

func fToC(f int) int {
	return (f - 32) * 5 / 9
}

func cToF(c int) int {
	return c*9/5 + 32
}

//flowdockMsg struct all the information we care about from flowdock message of type message
type flowdockMsg struct {
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

//flowdockUpdatedMsg struct all the information we care about from flowdock message of type update message
type flowdockUpdatedMsg struct {
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

//flowdockComment struct all the information we care about from flowdock message of type comment
type flowdockComment struct {
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

//flows struct that holds information about all the flows we can access
type flows struct {
	Id                 string
	Name               string
	Parameterized_name string
	Email              string
	Description        string
	Url                string
	Web_url            string
	Unread_mentions    int
}
