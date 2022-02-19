package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	//"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	//"os"
)

// check checks if there is an error and panics if there is one.
// Used to check for thrown configuration errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// contactAlpha Asks the alpha microservice the question.
// Takes the question as a map parameter and responds with a map answer.
func contactAlpha(query map[string]interface{}) (map[string]interface{}, error) {
	// Convert map to JSON string
	jsonBytes, err := json.Marshal(query)
	// Handle query format error
	if err != nil {
		return nil, err
	}
	// Set up request
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:3001/alpha", bytes.NewBuffer(jsonBytes))
	// Handle request set up error
	if err != nil {
		return nil, err
	}
	// Perform request
	rsp, err := client.Do(req)
	// Handle request error
	if err != nil {
		return nil, err
	}
	// Handle request not being successful
	if rsp.StatusCode != 200 {
		return nil, errors.New("alpha failure")
	}
	// Read response
	body, err := io.ReadAll(rsp.Body)
	// Handle response error
	if err != nil {
		return nil, err
	}
	u := map[string]interface{}{}
	err = json.Unmarshal(body, &u)
	// Handle response format error
	if err != nil {
		return nil, err
	}
	return u, nil
}

// contactTTS asks the TTS microservice to convert text to speech.
// Takes the text as a map parameter and responds with a map answer containing a base64 encoded answer.
func contactTTS(query map[string]interface{}) (map[string]interface{}, error) {
	// Convert map to JSON string
	jsonBytes, err := json.Marshal(query)
	// Handle query format error
	if err != nil {
		return nil, err
	}
	// Set up request
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:3003/tts", bytes.NewBuffer(jsonBytes))
	// Handle request set up error
	if err != nil {
		return nil, err
	}
	// Perform request
	rsp, err := client.Do(req)
	// Handle request error
	if err != nil {
		return nil, err
	}
	// Handle request not being successful
	if rsp.StatusCode != 200 {
		return nil, errors.New("tts failure")
	}
	// Read response
	body, err := io.ReadAll(rsp.Body)
	// Handle response error
	if err != nil {
		return nil, err
	}
	u := map[string]interface{}{}
	err = json.Unmarshal(body, &u)
	// Handle response format error
	if err != nil {
		return nil, err
	}
	return u, nil
}

// contactTTS asks the TTS microservice to convert text to speech.
// Takes the speech as a map parameter containing a base64 encoded value and responds with a map answer.
func contactSTT(query map[string]interface{}) (map[string]interface{}, error) {
	// Convert map to JSON string
	jsonBytes, err := json.Marshal(query)
	// Handle query format error
	if err != nil {
		return nil, err
	}
	// Set up request
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:3002/stt", bytes.NewBuffer(jsonBytes))
	// Handle request set up error
	if err != nil {
		return nil, err
	}
	// Perform request
	rsp, err := client.Do(req)
	// Handle request error
	if err != nil {
		return nil, err
	}
	// Handle request not being successful
	if rsp.StatusCode != 200 {
		return nil, errors.New("stt failure")
	}
	// Read response
	body, err := io.ReadAll(rsp.Body)
	// Handle response error
	if err != nil {
		return nil, err
	}
	u := map[string]interface{}{}
	err = json.Unmarshal(body, &u)
	// Handle response format error
	if err != nil {
		return nil, err
	}
	return u, nil
}

// handleReq handles any requests coming into the microservice.
// The parameters are a http.ResponseWriter for the response and a http.Request representing a request.
func handleReq(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	// Decode request
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		// Check for JSON validity
		_, present := t["speech"]
		if present {
			//Convert speech to text
			rsp, err := contactSTT(t)
			if err != nil {
				// Error handling for STT errors
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting speech to text"))
				check(err)
				return
			}
			// Ask question to alpha
			rsp, err = contactAlpha(rsp)
			if err != nil {
				// Error handling for Alpha errors
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting alpha"))
				check(err)
				return
			}
			// Convert text to speech
			rsp, err = contactTTS(rsp)
			if err != nil {
				// Error handling for TTS errors
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting text to speech"))
				check(err)
				return
			}
			// Write response
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(rsp)
			if err != nil {
				// Error handling if response could not be encoded
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Could not encode response"))
				check(err)
			}
		} else {
			// Error handling for incorrectly formatted JSON
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("JSON in incorrect format"))
			check(err)
		}
	} else {
		// Error handling for non JSON requests
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("Request not in JSON format"))
		check(err)
	}
}

func main() {
	// Router
	r := mux.NewRouter()
	// End point
	r.HandleFunc("/alexa", handleReq).Methods("POST")
	err := http.ListenAndServe(":3000", r)
	check(err)
}
