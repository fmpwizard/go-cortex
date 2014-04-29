package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//WitHandler is am http request handler that looks for the "q" query parameter
//and sends it to Wit for processing.
func WitHandler(w http.ResponseWriter, r *http.Request) {
	//read the "q" GET query parameter and pass it to
	// the wit service
	message := r.FormValue("q")
	if len(message) > 0 {
		ret := ProcessIntent(FetchIntent(message))
		//print what we understood from your request to the browser.
		msg := fmt.Sprintf("Turning light %v %s", ret.Arduino.Light, ret.Arduino.Action)
		fmt.Fprintf(w, msg)
	} else {
		fmt.Fprintf(w, "Please add a /wit?q=<text here> to the url")
	}
}

//FetchIntent is the whole go wit wrapper, if you call it that.
//We send the query string to wit, parse the result json
//into a struct and return it.
func FetchIntent(str string) WitMessage {

	url := "https://api.wit.ai/message?q=" + url.QueryEscape(str)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.WitAccessToken))
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Requesting wit's api gave: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode == 401 {
		log.Fatalln("Access denied, check your wit access token ")
	}

	return ProcessWitResponse(res.Body)

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
		return WitMessage{}, errors.New("no sound in file")
	}

	url := "https://api.wit.ai/speech"
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.WitAccessToken))
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

	return ProcessWitResponse(res.Body), nil

}

//ProcessWitResponse gets the raw response from the http request, and
//returns a WitMessage with all the information we got from Wit
func ProcessWitResponse(message io.ReadCloser) WitMessage {
	intent, _ := ioutil.ReadAll(message)

	jsonString := string(intent[:])
	_ = jsonString

	var jsonResponse WitMessage
	err := json.Unmarshal(intent, &jsonResponse)
	if err != nil {
		log.Println("error parsing json: ", err)
		log.Printf("plain text json was %+v", jsonString)
	}

	var numbers []WitNumber
	var number WitNumber

	err = json.Unmarshal(jsonResponse.Outcome.Entities.RawGithub, &numbers)
	if err != nil {
		//log.Println("1 error parsing number json: ", err)
		//log.Println("string number object is: ", string(jsonResponse.Outcome.Entities.RawGithub))
		err = json.Unmarshal(jsonResponse.Outcome.Entities.RawGithub, &number)
		if err != nil {
			//log.Println("2 error parsing number json: ", err)
		} else {
			jsonResponse.Outcome.Entities.MultipleNumber = []WitNumber{number}
		}

	} else {
		jsonResponse.Outcome.Entities.MultipleNumber = numbers
	}

	//log.Printf("a: %+v\n\n\n", jsonResponse)
	//log.Printf("b: %+v\n\n\n", jsonString)

	return jsonResponse

}

//ProcessIntent gets the json parsed result from wit.ai and
//depending on the intent, it calles the right service.
//So far we only have one service, the Arduino lights service
func ProcessIntent(jsonResponse WitMessage) WitResponse {
	switch jsonResponse.Outcome.Intent {
	case "lights":

		for _, row := range jsonResponse.Outcome.Entities.MultipleNumber {

			light := row.Value
			action := jsonResponse.Outcome.Entities.OnOff.Value
			Arduino(action, light)
			msg := fmt.Sprintf("Turning light %v %s", light, action)
			_ = msg
			return WitResponse{
				WitArduinoResponse{light, action},
				WitTemperatureResponse{},
				WitGithubResponse{},
			}
		}
	case "temperature":
		unit := jsonResponse.Outcome.Entities.Temperature.Value.Unit
		temperature := jsonResponse.Outcome.Entities.Temperature.Value.Temperature
		return WitResponse{
			WitArduinoResponse{},
			WitTemperatureResponse{unit, temperature},
			WitGithubResponse{},
		}
	case "github":
		var issues []int
		for _, row := range jsonResponse.Outcome.Entities.MultipleNumber {
			//log.Printf("1\n\n%+v\n\n", len(jsonResponse.Outcome.Entities.MultipleNumber))
			//log.Printf("2\n\n%+v\n\n", row)
			issues = append(issues, row.Value)
		}
		return WitResponse{
			WitArduinoResponse{},
			WitTemperatureResponse{},
			WitGithubResponse{issues},
		}

	}
	return WitResponse{}
}

//These make up the different parts of the wit result
//There are more options, but I'm using only these so far.

//WitMessage represents the payload we get from Wit as a response to processing
//the text or voice file we sent.
type WitMessage struct {
	MsgID   string `json:"msg_id"`
	MsgBody string `json:"msg_body"`
	Outcome WitMessageOutcome
}

//WitMessageOutcome gives you the Intent, the entities and a confidence value
//which you can use to discard or accept the intent we get from Wit.
//Lower than 0.8 means they are not very accurate.
type WitMessageOutcome struct {
	Intent     string
	Entities   WitMessageEntities
	Confidence float64
}

//WitMessageEntities contains all the possible entities we process from Wit
type WitMessageEntities struct {
	Location       WitLocation
	OnOff          WitOnOff
	RawGithub      json.RawMessage `json:"github_issue"`
	MultipleNumber []WitNumber
	SingleNumber   WitNumber `json:"number"`
	Temperature    WitTemperature
}

//WitLocation is the Location entity
type WitLocation struct {
	End       int
	Start     int
	Value     string
	Body      string
	Suggested bool
}

//WitOnOff is the on_off entity
type WitOnOff struct {
	Value string
}

//WitNumber is the wit/number entity
type WitNumber struct {
	End   int
	Start int
	Value int
	Body  string
}

//WitTemperature gives you the numeric value plus the units as a string
type WitTemperature struct {
	End   int
	Start int
	Value WitTemperatureValue
	Body  string
}

//WitTemperatureValue is the actual value and unit
type WitTemperatureValue struct {
	Unit        string
	Temperature int
}

//WitResponse holds just the information you need to act on each intent
type WitResponse struct {
	Arduino     WitArduinoResponse
	Temperature WitTemperatureResponse
	Github      WitGithubResponse
}

//WitArduinoResponse gives you the light number and a string representing on/off for the light number
type WitArduinoResponse struct {
	Light  int
	Action string
}

//WitTemperatureResponse gives you the Unit and degrees
type WitTemperatureResponse struct {
	Unit    string
	Degrees int
}

//WitGithubResponse gives you a slice of issue numbers
type WitGithubResponse struct {
	issues []int
}
