// Simple API wrapper for https://www.captionbot.ai/.
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

// Root path of Caption Bot URL.
// All requests will be paths starting from here.
var BASE_URL = "https://www.captionbot.ai/api/"

// Struct to hold data for API URL caption requests.
type CaptionBotRequest struct {
	ConversationId string `json:"conversationId"`
	UserMessage    string `json:"userMessage"`
	WaterMark      string `json:"waterMark"`
}

// Struct to hold data for API URL caption responses.
type CaptionBotResponse struct {
	ConversationId string
	UserMessage    string
	WaterMark      string
	Status         string
	BotMessages    []string
}

// Struct to hold "session" state.
// - conversationId: given during call to Initialize()
//                   Should be used for subsequent requests.
// - waterMark:      is updated per URL caption response.
// Note: consequences of not maintaining state is unknown.
type CaptionBotClientState struct {
	waterMark      string
	conversationId string
}

// Struct representing one session with CaptionBot.
type CaptionBot struct {
	state CaptionBotClientState
}

// Interface for methods for one CaptionBot session.
type CaptionBotConnection interface {
	URLCaption(url string) string
	Initialize() string
}

// POST request that starts a URL caption request on the server.
// Result will need to be retrieved by a subsequent GET request
// with the same parameters used here.
func CreateCaptionTask(data bytes.Buffer) {
	client := &http.Client{}
	queryURL := BASE_URL + "/message"
	req, _ := http.NewRequest("POST", queryURL, &data)
	req.Header.Add("Content-Type", "application/json; charset=utf8")
	resp, postErr := client.Do(req)
	defer resp.Body.Close()
	if postErr != nil {
		panic(postErr)
	}
}

// Create Values struct from state struct
func MakeValuesFromState(imgURL string, state CaptionBotClientState) url.Values {
	v := url.Values{}
	v.Set("conversationId", state.conversationId)
	v.Set("userMessage", imgURL)
	v.Set("waterMark", state.waterMark)
	return v
}

// Sanitize raw caption response from GET request.
// Currently, this method will:
// - remove starting and trailing double-quotes.
// - replace escaped double-quotes with double-quotes.
func SanitizeCaptionByteArray(data []byte) []byte {
	// Remove starting and trailing double-quotes
	trimmed := data[1 : len(data)-1]

	// Replace escaped double-quote with regular double-quote
	unescaped := strings.Replace(string(trimmed), "\\\"", "\"", -1)

	return []byte(unescaped)
}

// Sanitize caption string.
// Currently, this method will:
// - remove escaped newlines with newlines. 
func SanitizeCaptionString(caption string) string {
	// Replace escaped newlines with regular newlines
	return strings.Replace(caption, "\\n", "\n", -1)
}

// Send request to /init endpoint to retrieve conversationId.
// This is a session variable used in the state struct.
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

// Entry method for getting caption for image pointed to by URL.
// Performs a POST request to start the caption task.
// Then performs a GET request to retrieve the result.
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
	// 1. Trim() double-quotes from start and end.
	// 2. Remove escaped double-quotes in JSON.
	captionData := SanitizeCaptionByteArray(captionRawData)

	// Unmarshal it
	var captionJSON CaptionBotResponse
	captionJSONErr := json.Unmarshal(captionData, &captionJSON)
	if captionJSONErr != nil {
		panic(captionJSONErr)
	}

	// Update the state with the new watermark.
	// This is a side-effect.
	captionBot.state.waterMark = captionJSON.WaterMark

	//requestedURL := captionJSON.BotMessages[0]
	caption := captionJSON.BotMessages[1]

	return SanitizeCaptionString(caption)
}

