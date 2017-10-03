package captionbot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"bytes"
	"fmt"
	"strings"
)

var BASE_URL = "https://www.captionbot.ai/api/"

type CaptionBotRequest struct {
	ConversationId string `json:"conversationId"`
	UserMessage    string `json:"userMessage"`
	WaterMark      string `json:"waterMark"`
}

type CaptionBotResponse struct {
	ConversationId string
	UserMessage    string
	WaterMark      string
	Status         string
	BotMessages    []string
}

type CaptionBotClientState struct {
	waterMark      string
	conversationId string
}

type CaptionBot struct {
	state CaptionBotClientState
}

type CaptionBotConnection interface {
	URLCaption(url string) string
	Initialize() string
}

func CreateCaptionTask(data bytes.Buffer) {
	client := &http.Client{}
	queryURL := BASE_URL + "/message"
	req, _ := http.NewRequest("POST", queryURL, &data)
	req.Header.Add("Content-Type", "application/json; charset=utf8")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func MakeValuesFromState(imgURL string, state CaptionBotClientState) url.Values {
	v := url.Values{}
	v.Set("conversationId", state.conversationId)
	v.Set("userMessage", imgURL)
	v.Set("waterMark", state.waterMark)
	return v
}

func SanitizeCaptionRawData(data []byte) []byte {
	// Remove starting and trailing double-quotes
	trimmed := data[1 : len(data)-1]

	// Replace escaped double-quote with regular double-quote
	unescaped := strings.Replace(string(trimmed), "\\\"", "\"", -1)

	// Replace escaped newlines with regular newlines
	unescaped = strings.Replace(unescaped, "\\\\n", " ", -1)

	return []byte(unescaped)
}

func (captionBot *CaptionBot) Initialize() {
	resp, getErr := http.Get(BASE_URL + "init")
	defer resp.Body.Close()
	if getErr != nil {
		panic(getErr)
	}

	bodyByteArray, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		panic(bodyErr)
	}

	captionBot.state.conversationId = strings.Trim(string(bodyByteArray[:]), "\"")
}

func (captionBot *CaptionBot) URLCaption(url string) string {
	if captionBot.state.conversationId == "" {
		fmt.Println(
			"CaptionBot not initialize.",
			"Please call CaptionBot::Initialize().",
		)
		return ""
	}

	// Create JSON data from state for POST request
	requestData := CaptionBotRequest{
		ConversationId: captionBot.state.conversationId,
		UserMessage:    url,
		WaterMark:      captionBot.state.waterMark,
	}
	jsonData, marshalErr := json.Marshal(requestData)
	if marshalErr != nil {
		panic(marshalErr)
	}

	/*
	  - This request kicks off a caption task on the server for 
	    the picture identified by `requestData.UserMessage`
	  - the result will need to be retrieved with a subseqent
	    GET request using the above data as URL-encoded params.
	*/
	var data bytes.Buffer
	data.Write(jsonData)
	CreateCaptionTask(data)

	// Create Values struct for URL encoded params
	v := MakeValuesFromState(url, captionBot.state)

	// Actually Query for Caption
	queryURL := BASE_URL + "/message"
	resp, getErr := http.Get(queryURL + "?" + v.Encode())
	defer resp.Body.Close()
	if getErr != nil {
		panic(getErr)
	}

	captionRawData, readBodyErr := ioutil.ReadAll(resp.Body)
	if readBodyErr != nil {
		panic(readBodyErr)
	}

	// We need to format the returned string so Golang can unmarshal it.
	// 1. Trim() double-quotes from start and end
	// 2. Remove escaped double-quotes in JSON
	captionData := SanitizeCaptionRawData(captionRawData)

	// Unmarshal it
	var captionJSON CaptionBotResponse
	captionJSONErr := json.Unmarshal(captionData, &captionJSON)
	if captionJSONErr != nil {
		panic(captionJSONErr)
	}

	captionBot.state.waterMark = captionJSON.WaterMark

	//requestedURL := captionJSON.BotMessages[0]
	caption := captionJSON.BotMessages[1]
	return caption
}

