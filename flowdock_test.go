package main

import (
	"bytes"
	"testing"
)

func TestFlowsParse(t *testing.T) {
	parseAvailableFlows(NopCloser{NopCloser{bytes.NewBufferString(mockedFlows)}})
	if len(availableFlows) != 2 {
		t.Errorf("parseAvailableFlows didn't parse payload, gave %+v", availableFlows)
	}
}

func TestGetFlowUrl(t *testing.T) {
	url, _ := getFlowURL("aaaaaaaa-d97b-0000-1111-555598671f8c")
	if url != "https://api.flowdock.com/flows/fmpwizard/huston" {
		t.Errorf("getFlowUrl gave %+v", url)
	}
}

func TestGetFlowName(t *testing.T) {
	name, _ := getFlowName("aaaaaaaa-d97b-0000-1111-555598671f8c")
	if name != "huston" {
		t.Errorf("getFlowName gave %+v", name)
	}
}

func TestParseUsers(t *testing.T) {
	parseUsers()([]byte(mockedUsers))
	if len(currentUsers) != 2 {
		t.Errorf("Didn't get two users, got: %+v", len(currentUsers))
	}
}

const mockedFlows = (`[
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

const mockedUsers = (`
[
    {
        "id": 4877,
        "nick": "cortex",
        "email": "diego+cortex@fmpwizard.com",
        "avatar": "https://d2cxspbh1aoie1.cloudfront.net/avatars/local/546440c99fa5e511111111111222222222222225/",
        "name": "Cortex",
        "website": null
    },
    {
        "id": 31347,
        "nick": "Diego",
        "email": "fmpwizard@gmail.com",
        "avatar": "https://d2cxspbh1aoie1.cloudfront.net/avatars/5fcd8164c5ae83f10f060788840f258e/",
        "name": "Diego Medina",
        "website": ""
    }
]
`)
