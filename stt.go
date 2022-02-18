package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
)

type Keys struct {
	Speech string `json:"Speech"`
	Alpha  string `json:"Alpha"`
}

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
)

var keys Keys

// handleRequest Handles the request.
// w is the ResponseWriter, r is a pointer to the request.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Decode request
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		// Decode speech
		speech, present := t["speech"].(string)
		if present {
			data, err := base64.StdEncoding.DecodeString(speech)
			if err == nil {
				// Ask microsoft what is being said
				text, err := handleResponse(data)
				if err == nil {
					// Encode Response
					u := map[string]interface{}{"text": text}
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(u)
					if err != nil {
						// Handle error when encoding response
						w.WriteHeader(http.StatusInternalServerError)
						_, err := w.Write([]byte("Something went wrong encoding the response"))
						check(err)
					}
				} else {
					// Handle error contacting microsoft
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Something went wrong contacting the STT provider"))
					check(err)
				}
			} else {
				// Handle speech not being encoded correctly
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("Speech is not a Base 64 encoded string"))
				check(err)
			}
		} else{
			// Handle invalid JSON format
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("Incorrect JSON format"))
			check(err)
		}
	} else {
		// Handle non JSON format
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Request format is not JSON"))
		check(err)
	}

}

// handleResponse Handles contacting the microsoft API.
// Takes base64 encoded speech fill as a parameter and returns a string of the speech.
func handleResponse(speech []byte) (string, error) {
	// Set up request
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	check(err)
	// Set headers
	req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	// Perform request
	rsp, err := client.Do(req)
	// If response code is not 200 raise an error
	if rsp.StatusCode != 200 {
		return "", errors.New("microsoft didn't like that")
	}
	// Read body
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	// Decode response
	t := map[string]interface{}{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&t)
	check(err)
	// Get text
	text := t["DisplayText"].(string)
	return text, err
}

// check checks if there is a configuration error. Panics if necessary
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Load Keys
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	err = json.Unmarshal(byteValue, &keys)
	check(err)
	// Router
	r := mux.NewRouter()
	// Endpoint
	r.HandleFunc("/stt", handleRequest).Methods("POST")
	err = http.ListenAndServe(":3002", r)
	check(err)
}
