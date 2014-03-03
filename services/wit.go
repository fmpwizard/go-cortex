package services

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var witAccessToken string

func init() {
	flag.StringVar(&witAccessToken, "witAccessToken", "", "Your WIT access token")
}

func FetchIntent(str string) WitMessage {

	url := "https://api.wit.ai/message?q=" + url.QueryEscape(str)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", witAccessToken))
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Requesting wit's api gave: %v", err)
	}
	if res.StatusCode == 401 {
		log.Fatalln("Access denied, check your wit access token ")
	}

	intent, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	jsonString := string(intent[:])
	_ = jsonString

	var jsonResponse WitMessage
	err = json.Unmarshal(intent, &jsonResponse)
	if err != nil {
		log.Println("error parsing json: ", err)
	}

	//log.Printf("%+v\n\n\n", jsonResponse)
	//log.Printf("%+v\n\n\n", jsonString)

	return jsonResponse

}

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
