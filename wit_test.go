package main

import (
	"bytes"
	"io"
	"testing"
)

type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error { return nil }

func TestProcessWitResponseGithubMultipleIssues(t *testing.T) {

	withJSON := stringToReadeClosser(githubMultipleIssues)
	numbers := ProcessWitResponse(withJSON).Outcome.Entities.MultipleNumber

	if len(numbers) != 2 {
		t.Errorf("ProcessWitResponse didn't parse the 'numbers' array. We got %+v\n", numbers)
	}
}

func TestProcessWitResponseGithubSingleIssue(t *testing.T) {

	withJSON := stringToReadeClosser(githubSingleIssue)
	number := ProcessWitResponse(withJSON).Outcome.Entities.MultipleNumber

	if len(number) != 1 {
		t.Errorf("ProcessWitResponse didn't parse the 'number' object. We got %+v\n", number)
	}

	if number[0].Value != 45 {
		t.Errorf("Head item's Value is not 45. We got %+v\n", number)
	}
}

func TestProcessWitResponseSingleLight(t *testing.T) {

	withJSON := stringToReadeClosser(lightPayload)
	number := ProcessWitResponse(withJSON).Outcome.Entities.SingleNumber

	if number.Value != 1 {
		t.Errorf("ProcessWitResponse didn't parse the 'number' object. We got %+v\n", number)
	}

}

func stringToReadeClosser(s string) io.ReadCloser {
	return NopCloser{bytes.NewBufferString(s)}
}

const githubMultipleIssues = `{
  "msg_body": "look at #45, #102 and work on those",
  "outcome": {
    "intent": "github",
    "confidence": 0.997,
    "entities": {
      "github_issue": [
        {
          "value": 45,
          "body": "45,",
          "start": 9,
          "end": 11
        },
        {
          "value": 102,
          "body": "102 ",
          "start": 14,
          "end": 17
        }
      ]
    }
  }
}`

const githubSingleIssue = `{
  "msg_body": "look at #45 and work on those",
  "outcome": {
    "intent": "github",
    "confidence": 0.997,
    "entities": {
      "github_issue":
        {
          "value": 45,
          "body": "45,",
          "start": 9,
          "end": 11
        }
    }
  }
}`

const lightPayload = `{
  "msg_body": "turn the light one on please",
  "outcome": {
    "intent": "lights",
    "confidence": 1,
    "entities": {
      "on_off": {
        "value": "on"
      },
      "number": {
        "value": 1,
        "body": "one ",
        "start": 15,
        "end": 18
      }
    }
  }
}`
