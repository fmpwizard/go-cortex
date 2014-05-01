package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var availableFlows []flows
var currentUsers []user

func tokenFlowdock() string {
	return base64.StdEncoding.EncodeToString([]byte(config.FlowdockAccessToken))
}

//ListenStream starts pulling the flowdock stream api
func listenStream() {
	fetchFlows()
	go fetchUserSchedule()
	res := connectToFlow()
	defer res.Body.Close()
	for {
		processFlowRow(parseFlowRow(bufio.NewReader(res.Body)))
	}
}

//fetchFlows fetches all the flows we have access to
func fetchFlows() {
	performGet("flows", parseAvailableFlows())
}

func parseAvailableFlows() parseCallback {
	return func(payload []byte) {
		err := json.Unmarshal(payload, &availableFlows)
		if err != nil {
			log.Fatalf("Error parsing flows data %+v", err)
		}
	}
}

func connectToFlow() *http.Response {
	url := fmt.Sprintf("https://%+v@stream.flowdock.com/flows?filter=%s", config.FlowdockAccessToken, config.Flows)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", tokenFlowdock()))
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
		log.Fatalf("something went wrong reading the payload: %s", err)
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
		intent, err := FetchIntent(flowMessage.Content)
		if err != nil {
			witErrorResponseToFlowdock(flowMessage, err)
		} else {
			ret := ProcessIntent(intent)
			replyToFlow(ret, flowMessage.Id, flowMessage.Flow)
		}

	case "message-edit":
		json.Unmarshal(line, &flowUpdatedMessage)
		intent, err := FetchIntent(flowUpdatedMessage.Content.Updated_content)
		if err != nil {
			witErrorResponseToFlowdock(flowMessage, err)
		} else {
			ret := ProcessIntent(intent)
			replyToFlow(ret, flowUpdatedMessage.Id, flowUpdatedMessage.Flow)
		}

	case "comment":
		if flowMessage.User != "77156" {
			var parentMessageID int64
			json.Unmarshal(line, &flowComment)
			intent, err := FetchIntent(flowComment.Content.Text)
			if err != nil {
				flowMessage.Id = parentMessageID
				witErrorResponseToFlowdock(flowMessage, err)
			} else {
				ret := ProcessIntent(intent)
				for _, v := range flowComment.Tags {
					if strings.Contains(v, "influx") {
						parentID, _ := strconv.ParseInt(strings.Split(v, ":")[1], 0, 64)
						parentMessageID = parentID
					}
				}
				replyToFlow(ret, parentMessageID, flowComment.Flow)
			}

		} else {
			//log.Println("skipping Cortex's message.")
		}
	}
}

func witErrorResponseToFlowdock(flowMessage flowdockMsg, err error) {
	ret :=
		WitResponse{
			WitArduinoResponse{},
			WitTemperatureResponse{},
			WitGithubResponse{},
			witError{fmt.Sprintf("Error: %+v", err)},
		}
	replyToFlow(ret, flowMessage.Id, flowMessage.Flow)
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
	} else if ret.Error.msg != "" {
		flowdockPost(ret.Error.msg, originalMessageID, flowID)
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
	client := &http.Client{}
	payload := []byte(`{
	  "event": "comment",
	  "content":"` + message + `"}`)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", tokenFlowdock()))
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

func fetchUserSchedule() {
	for _ = range time.Tick(1 * time.Minute) {
		fetchUsers()
	}

}

func fetchUsers() {
	performGet("users", parseUsers())
}

func performGet(path string, f parseCallback) {
	url := fmt.Sprintf("https://%+v@api.flowdock.com/%+v", config.FlowdockAccessToken, path)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error getting %+v: %v", path, err)
	} else if res.StatusCode != 200 {
		log.Fatalf("got status code %+v", res.StatusCode)
	}
	dataAsJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("error reading body, got: %+v", err)
	}
	f([]byte(dataAsJson))
	res.Body.Close()
}

func parseUsers() parseCallback {
	return func(payload []byte) {
		err := json.Unmarshal(payload, &currentUsers)
		if err != nil {
			log.Printf("Unabled to parse users list, got %+v", err)
		}
		return
	}
}

func getCortexUserID(email string) int64 {
	for _, value := range currentUsers {
		if value.Email == email {
			return value.ID
		}
	}
	return 0
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

//we fetch all the users currently logged in and use this struct to stuff them into
type user struct {
	ID      int64
	Nick    string
	Email   string
	Avatar  string
	Mame    string
	Website string
}

type parseCallback func([]byte)
