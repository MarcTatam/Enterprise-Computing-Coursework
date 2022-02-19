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
// Takes the question as a map argument and responds with a map answer.
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
// Takes the question as a map argument and responds with a map answer.
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
// Takes the question as a map argument and responds with a map answer.
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

func handleReq(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	// Decode request
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		_, present := t["speech"]
		if present {
			rsp, err := contactSTT(t)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting speech to text"))
				check(err)
				return
			}
			rsp, err = contactAlpha(rsp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting alpha"))
				check(err)
				return
			}
			rsp, err = contactTTS(rsp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Error contacting text to speech"))
				check(err)
				return
			}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(rsp)
			if err != nil {
				// Error handling if response could not be encoded
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Could not encode response"))
				check(err)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("JSON in incorrect format"))
			check(err)
		}
	} else {
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
	err = http.ListenAndServe(":3000", r)
	check(err)
}
