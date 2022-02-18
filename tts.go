package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
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
	URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
)

var keys Keys

// handleRequest Handles the request made to the RESTful api.
// It takes two parameters,w the response writer and r the pointer to the request
// This function will only have an error if there is some internal issue with the configuration
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Decode Request
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		// Get text from request
		text, present := t["text"].(string)
		if present {
			// Make request to Microsoft API
			speech, err := handleResponse(xmlFormat(text))
			if err == nil {
				// Encode Response
				u := map[string]interface{}{"speech": speech}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					// Handle errors in response encoding
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Could not encode JSON"))
					check(err)
				}
			} else {
				// Handle errors in contacting the Microsoft API
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Something went wrong contacting the text to speech api"))
				check(err)
			}
		} else {
			// Handle the JSON not being in the correct format
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("Incorrect JSON format"))
			check(err)
		}
	} else {
		// Handle request body not being in JSON format
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Request not in JSON format"))
		check(err)
	}
}

// xmlFormat formats the speech into xml format for the microsoft API.
// Takes a string to format as a parameter and returns a string representing a SpeechXML document
func xmlFormat(text string) string {
	return "<?xml version=\"1.0\"?><speak version=\"1.0\" xml:lang=\"en-US\"><voice xml:lang=\"en-US\" name=\"en-US-JennyNeural\">" + text + "</voice></speak>"
}

// handleResponse makes the request to the Microsoft API for text to speech.
// Takes a string and returns the base64 encoded speech file
func handleResponse(text string) ([]byte, error) {
	// Set up the request
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewBuffer([]byte(text)))
	if err != nil {
		return nil, err
	}
	// Set request headers
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	//Send off request
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	// If the response code is not 200 raise an error
	if rsp.StatusCode != 200 {
		fmt.Println(rsp.StatusCode)
		return nil, errors.New("microsoft didn't like that")
	}
	return b, nil
}

// check checks if there is an error and panics if there is one.
// Used to check for thrown configuration errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// main contains router
func main() {
	// Load keys from json file
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	check(err)
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	err = json.Unmarshal(byteValue, &keys)
	check(err)

	// Router
	r := mux.NewRouter()
	// Endpoint
	r.HandleFunc("/tts", handleRequest).Methods("POST")
	err = http.ListenAndServe(":3003", r)
	check(err)
}
