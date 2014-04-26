package main

import (
	"bytes"
	"testing"
)

func TestFlowsParse(t *testing.T) {
	parsedFlows := FlowsParse(NopCloser{NopCloser{bytes.NewBufferString(MockedFlows)}})
	if len(parsedFlows) != 2 {
		t.Errorf("FlowsParse didn't parse payload, gave %+v", parsedFlows)
	}
}

func TestGetFlowUrl(t *testing.T) {
	url, _ := GetFlowURL("aaaaaaaa-d97b-0000-1111-555598671f8c")
	if url != "https://api.flowdock.com/flows/fmpwizard/huston" {
		t.Errorf("GetFlowUrl gave %+v", url)
	}
}

func TestGetFlowName(t *testing.T) {
	name, _ := GetFlowName("aaaaaaaa-d97b-0000-1111-555598671f8c")
	if name != "huston" {
		t.Errorf("GetFlowName gave %+v", name)
	}
}

const MockedFlows = (`[
    {
        "id": "aaaaaaaa-d97b-0000-1111-555598671f8c",
        "name": "Huston",
        "parameterized_name": "huston",
        "email": "huston@fmpwizard.flowdock.com",
        "description": "Me",
        "url": "https://api.flowdock.com/flows/fmpwizard/huston",
        "web_url": "https://www.flowdock.com/app/fmpwizard/huston",
        "access_mode": "organization",
        "api_token": "123456789",
        "open": true,
        "joined": true,
        "last_message_at": "2014-04-02T12:59:32.187Z",
        "unread_mentions": 0,
        "organization": {
            "active": true,
            "id": 30602,
            "name": "FMPwizard",
            "parameterized_name": "fmpwizard",
            "url": "https://api.flowdock.com/organizations/fmpwizard",
            "user_count": 2,
            "user_limit": 0
        }
    },
    {
        "id": "9ad75588-d97b-0000-1111-555598671f8c",
        "name": "Mission control",
        "parameterized_name": "mission-control",
        "email": "mission-control@fmpwizard.flowdock.com",
        "description": "Me",
        "url": "https://api.flowdock.com/flows/fmpwizard/mission-control",
        "web_url": "https://www.flowdock.com/app/fmpwizard/mission-control",
        "access_mode": "organization",
        "api_token": "123456789",
        "open": true,
        "joined": true,
        "last_message_at": "2014-04-02T12:59:32.187Z",
        "unread_mentions": 0,
        "organization": {
            "active": true,
            "id": 30602,
            "name": "FMPwizard",
            "parameterized_name": "fmpwizard",
            "url": "https://api.flowdock.com/organizations/fmpwizard",
            "user_count": 2,
            "user_limit": 0
        }
    }
]`)
