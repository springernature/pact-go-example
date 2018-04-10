package main

import (
	"log"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"bytes"
	"fmt"
	"github.com/pact-foundation/pact-go/types"
	"path/filepath"
	"os"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)

func TestConsumer(t *testing.T) {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Port:     6666, // Ensure this port matches the daemon port!
		Consumer: "Consumer",
		Provider: "Provider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		var jsonStr = []byte(`{"s":"hello, world"}`)
		u := fmt.Sprintf("http://localhost:%d/uppercase", pact.Server.Port)
		req, err := http.NewRequest("POST", u, bytes.NewBuffer(jsonStr))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		if _, err = http.DefaultClient.Do(req); err != nil {
			return err
		}

		return err
	}

	body := dsl.Like(fmt.Sprintf(`{"v":"%s"}`, "HELLO, WORLD"))

	// Set up our expected interactions.
	pact.
		AddInteraction().
		UponReceiving("A request with a string").
		WithRequest(dsl.Request{
		Method:  "POST",
		Path:    "/uppercase",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"s":"hello, world"}`,
	}).
		WillRespondWith(dsl.Response{
		Status:  200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    body,
	})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

	// Publish the Pacts...
	p := dsl.Publisher{}
	err := p.Publish(types.PublishRequest{
		PactURLs:        []string{filepath.FromSlash(fmt.Sprintf("%s/consumer-provider.json", pactDir))},
		PactBroker:      "http://localhost",
		ConsumerVersion: "2.2.2",
		Tags:            []string{"latest", "stable"},
		BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
	})

	if err != nil {
		log.Println("ERROR: ", err)
	}
}
