package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	waterMark string
	conversationId string
}

func GetConversationID() string {
	resp, getErr := http.Get(BASE_URL + "init")
	if getErr != nil {
		fmt.Println("error:", getErr)
	}
	defer resp.Body.Close()
	bodyByteArray, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		fmt.Println("error:", bodyErr)
	}

	return strings.Trim(string(bodyByteArray[:]), "\"")
}

func PrimeCaptionBot(data bytes.Buffer) {
	client := &http.Client{}
	queryURL := BASE_URL + "/message"
	req, _ := http.NewRequest("POST", queryURL, &data)
	req.Header.Add("Content-Type", "application/json; charset=utf8")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	//caption, err := ioutil.ReadAll(resp.Body)
	//fmt.Println("Caption:", caption)
}

func MakeValuesFromRequestData(requestData CaptionBotRequest) url.Values {
	v := url.Values{}
	v.Set("conversationId", requestData.ConversationId)
	v.Set("userMessage", requestData.UserMessage)
	v.Set("waterMark", requestData.WaterMark)
	return v
}

func SanitizeCaptionRawData(data []byte) []byte {
	// Remove starting and trailing double-quotes
	trimmed := data[1:len(data)-1]

	// Replace escaped double-quote with regular double-quote
        unescaped := strings.Replace(string(trimmed), "\\\"", "\"", -1)

	// Replace escaped newlines with regular newlines
	unescaped = strings.Replace(unescaped, "\\\\n", " ", -1)

	return []byte(unescaped)
}

func URLCaption(url string, state CaptionBotClientState) string {
	// Create data for POST request
	requestData := CaptionBotRequest{
		ConversationId: state.conversationId,
		UserMessage:    url,
		WaterMark:      state.waterMark,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("JSON data: %s\n", jsonData)
	var data bytes.Buffer
	data.Write(jsonData)

	// Why is this necessary?
	PrimeCaptionBot(data)

	// Create Values struct for URL encoded params
	v := MakeValuesFromRequestData(requestData)

	// Actually Query for Caption
	queryURL := BASE_URL + "/message"
	resp, err := http.Get(queryURL + "?" + v.Encode())
	defer resp.Body.Close()
	captionRawData, _ := ioutil.ReadAll(resp.Body)

	// We need to format the returned string so Golang can unmarshal it.
	// 1. Trim() double-quotes from start and end
	// 2. Remove escaped double-quotes in JSON
	captionData := SanitizeCaptionRawData(captionRawData)
	fmt.Printf("JSON Response: %s\n", captionData)

	// Unmarshal it
	var captionJSON CaptionBotResponse
	captionJSONErr := json.Unmarshal(captionData, &captionJSON)
	if captionJSONErr != nil {
		fmt.Println("error:", captionJSONErr)
	}

	state.waterMark = captionJSON.WaterMark

	//requestedURL := captionJSON.BotMessages[0]
	caption := captionJSON.BotMessages[1]
	return caption
}

func main() {
	var state CaptionBotClientState
	state.conversationId = GetConversationID()

	var imgURL = "http://www.nhatqbui.com/assets/me.jpg"
	fmt.Println(URLCaption(imgURL, state))
}
