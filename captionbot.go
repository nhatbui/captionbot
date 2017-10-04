// Simple API wrapper for https://www.captionbot.ai/.
package captionbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
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
// 1) conversationId: given during call to Initialize()
//                    Should be used for subsequent requests.
// 2) waterMark:      is updated per URL caption response.
// (Note: consequences of not maintaining state is unknown.)
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
	URLCaption(url string) (string, error)
	Initialize() (string, error)
}

func New() (*CaptionBot, error) {
	var err error
	cb := &CaptionBot{}
	err = cb.Initialize()
	if err != nil {
		return cb, err
	}

	return cb, nil
}

// POST request that starts a URL caption request on the server.
// Result will need to be retrieved by a subsequent GET request
// with the same parameters used here.
func CreateCaptionTask(data bytes.Buffer) error {
	queryURL := BASE_URL + "/message"
	req, err := http.NewRequest("POST", queryURL, &data)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf8")
	resp, postErr := http.DefaultClient.Do(req)
	if postErr != nil {
		return postErr
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Non 2XX status code when POST-ing caption task.")
	}

	return nil
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
// 1) remove starting and trailing double-quotes.
// 2) replace escaped double-quotes with double-quotes.
func SanitizeCaptionByteArray(data []byte) []byte {
	// Remove starting and trailing double-quotes
	trimmed := data[1 : len(data)-1]

	// Replace escaped double-quote with regular double-quote
	unescaped := strings.Replace(string(trimmed), "\\\"", "\"", -1)

	return []byte(unescaped)
}

// Sanitize caption string.
// Currently, this method will:
// 1) remove escaped newlines with newlines.
func SanitizeCaptionString(caption string) string {
	// Replace escaped newlines with regular newlines
	return strings.Replace(caption, "\\n", "\n", -1)
}

// Send request to /init endpoint to retrieve conversationId.
// This is a session variable used in the state struct.
func (captionBot *CaptionBot) Initialize() error {
	resp, getErr := http.Get(BASE_URL + "init")
	if getErr != nil {
		return getErr
	}
	defer resp.Body.Close()

	bodyByteArray, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return bodyErr
	}

	captionBot.state.conversationId = strings.Trim(string(bodyByteArray[:]), "\"")
	return nil
}

// Entry method for getting caption for image pointed to by URL.
// Performs a POST request to start the caption task.
// Then performs a GET request to retrieve the result.
func (captionBot *CaptionBot) URLCaption(url string) (string, error) {
	var err error

	if captionBot.state.conversationId == "" {
		return "", fmt.Errorf(`CaptionBot not initialize.\n
                              Please call CaptionBot::Initialize().`)
	}

	// Create JSON data from state for POST request
	requestData := CaptionBotRequest{
		ConversationId: captionBot.state.conversationId,
		UserMessage:    url,
		WaterMark:      captionBot.state.waterMark,
	}

	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(requestData); err != nil {
		return "", err
	}

	/*
	  - This request kicks off a caption task on the server for
	    the picture identified by `requestData.UserMessage`
	  - the result will need to be retrieved with a subseqent
	    GET request using the above data as URL-encoded params.
	*/
	if err = CreateCaptionTask(data); err != nil {
		return "", err
	}

	// Create Values struct for URL encoded params
	v := MakeValuesFromState(url, captionBot.state)

	// Actually Query for Caption
	queryURL := BASE_URL + "/message"
	resp, getErr := http.Get(queryURL + "?" + v.Encode())
	if getErr != nil {
		return "", getErr
	}
	defer resp.Body.Close()

	// they return a json as string; unmarshal it into a string first then into caption bot response type
	var response string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	// Unmarshal it
	var captionJSON CaptionBotResponse
	captionJSONErr := json.Unmarshal([]byte(response), &captionJSON)
	if captionJSONErr != nil {
		return "", captionJSONErr
	}

	// Update the state with the new watermark.
	// This is a side-effect.
	captionBot.state.waterMark = captionJSON.WaterMark

	//requestedURL := captionJSON.BotMessages[0]
	caption := captionJSON.BotMessages[1]

	return caption, nil
}

// UploadCaption uploads a file and runs URLCaption on the result
func (captionBot *CaptionBot) UploadCaption(fileName string) (string, error) {
	// Make sure file exist, that its readable and then read it into memory
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	file.Close()

	// Prepare the post
	mimetype := mime.TypeByExtension(filepath.Ext(fileName))

	postbody := new(bytes.Buffer)
	writer := multipart.NewWriter(postbody)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filepath.Base(fileName)))
	h.Set("Content-Type", mimetype)
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}

	// Write the content
	part.Write(fileContents)

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%supload", BASE_URL), postbody)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Sanitize reply and return it
	urlCaption, err := captionBot.URLCaption(string(SanitizeCaptionByteArray(body)))
	if err != nil {
		return "", err
	}

	return urlCaption, nil
}
