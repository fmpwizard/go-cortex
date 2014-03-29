package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var witAccessToken string

//You need to get a wit access token to use their services
func init() {
	flag.StringVar(&witAccessToken, "witAccessToken", "", "Your WIT access token")
}

//FetchIntent is the whole go wit wrapper, if you call it that.
//We send the query string to wit, parse the result json
//into a struct and return it.
func FetchIntent(str string) WitMessage {

	url := "https://api.wit.ai/message?q=" + url.QueryEscape(str)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", witAccessToken))
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		log.Fatalf("Requesting wit's api gave: %v", err)
	}
	if res.StatusCode == 401 {
		log.Fatalln("Access denied, check your wit access token ")
	}

	return processWitResponse(res.Body)

}

//FetchVoiceIntent is like FetchIntent, but sends a wav file
// to the speech endpoint, Wit extracts the text from the sound file
//and then returns a json response with all the info we need.
func FetchVoiceIntent(filePath string) (WitMessage, error) {
	log.Println("reading file")
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("error: %v reading file", err)
	}
	if len(body) == 0 {
		return WitMessage{}, errors.New("No sound in file")
	}

	url := "https://api.wit.ai/speech"
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", witAccessToken))
	req.Header.Add("Content-Type", "audio/wav")
	log.Println("sending request")
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		log.Fatalf("Requesting wit's api gave: %v", err)
	}
	if res.StatusCode == 401 {
		log.Fatalln("Access denied, check your wit access token ")
	}

	return processWitResponse(res.Body), nil

}

//processWitResponse gets the raw response from the http request, and
//returns a WitMessage with all the information we got from Wit
func processWitResponse(message io.ReadCloser) WitMessage {
	intent, _ := ioutil.ReadAll(message)

	jsonString := string(intent[:])
	_ = jsonString

	var jsonResponse WitMessage
	err := json.Unmarshal(intent, &jsonResponse)
	if err != nil {
		log.Println("error parsing json: ", err)
	}

	log.Printf("%+v\n\n\n", jsonResponse)
	log.Printf("%+v\n\n\n", jsonString)

	return jsonResponse

}

//ProcessIntent gets the json parsed result from wit.ai and
//depending on the intent, it calles the right service.
//So far we only have one service, the Arduino lights service
func ProcessIntent(jsonResponse WitMessage) string {
	switch jsonResponse.Outcome.Intent {
	case "lights":
		light := jsonResponse.Outcome.Entities.Number.Value
		action := jsonResponse.Outcome.Entities.OnOff.Value
		Arduino(action, light)
		return fmt.Sprintf("Turning light %v %s", light, action)
	}
	return ""
}

//These make up the different parts of the wit result
//There are more options, but I'm using only these so far.

type WitMessage struct {
	MsgId   string            `json:"msg_id"`
	MsgBody string            `json:"msg_body"`
	Outcome WitMessageOutcome `json:"outcome"`
}

type WitMessageOutcome struct {
	Intent     string             `json:"intent"`
	Entities   WitMessageEntities `json:"entities"`
	Confidence float64            `json:"confidence"`
}

type WitMessageEntities struct {
	Location WitLocation `json:"location"`
	OnOff    WitOnOff    `json:"on_off"`
	Number   WitNumber   `json:"number"`
}

type WitLocation struct {
	End       int    `json:"end"`
	Start     int    `json:"start"`
	Value     string `json:"value"`
	Body      string `json:"body"`
	Suggested bool   `json:"suggested"`
}

type WitOnOff struct {
	Value string `json:"value"`
}

type WitNumber struct {
	End   int    `json:"end"`
	Start int    `json:"start"`
	Value int    `json:"value"`
	Body  string `json:"body"`
}
