package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"
)

type SlackRequestBody struct {
	Text string `json:"text"`
}

type Configuration struct {
	Frequency   int
	WebhookUrl  string
	SearchUrl   string
	SearchText  string
	DomSearch   string
	FoundMsg    string
	NotFoundMsg string
}

// Author: Edd Turtle
//	https://golangcode.com/send-slack-messages-without-a-library/
// SendSlackNotification will post to an 'Incoming Webook' url setup in
// Slack Apps. It accepts some text and the slack channel is saved within Slack.
func SendSlackNotification(webhookUrl string, msg string) error {

	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}

func retrievePage(searchUrl string) *goquery.Document {
	// make client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// make request
	response, err := client.Get("https://jobs.dev.to/")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	return document
}

func SendNotification(searchUrl string) {
	webhookUrl := viper.GetString("webhookUrl")
	changeMessage := viper.GetString("changeDetectedMsg") + " " + searchUrl
	SendSlackNotification(webhookUrl, changeMessage)
	log.Println(changeMessage)
}

func checkForListings(configuration Configuration, foundChange bool) bool {
	searchUrl := configuration.SearchUrl
	document := retrievePage(searchUrl)
	changeDetected := foundChange
	openings := document.Find(configuration.DomSearch)

	if strings.TrimSpace(openings.Text()) == configuration.SearchText {
		log.Println("No changes detected.")
		changeDetected = false
	} else {
		if foundChange == false {
			SendNotification(searchUrl)
			changeDetected = true
		}
	}

	return changeDetected
}

func configSetup() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/page-watcher/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func main() {

	var configuration Configuration
	configSetup()
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	foundChange := false
	for {
		foundChange = checkForListings(configuration, foundChange)
		if configuration.Frequency < 1 {
			configuration.Frequency = 1
		}
		time.Sleep(time.Duration(configuration.Frequency) * time.Minute)
	}
}
