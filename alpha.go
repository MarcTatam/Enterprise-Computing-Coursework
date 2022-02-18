package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const URI = "http://api.wolframalpha.com/v1/spoken"

type Keys struct {
	Speech string `json:"Speech"`
	Alpha  string `json:"Alpha"`
}

var keys Keys

// check checks if an error exists and panics if there is one
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// AlphaReq makes a request to Wolfram API.
// It takes a string as a question and responds with a string as a response
func AlphaReq(question string) (string, error) {
	// Set up HTTP request
	client := &http.Client{}
	uri := URI + "?appid=" + keys.Alpha + "&i=" + url.QueryEscape(question)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	// Perform request
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// If response code is not 200 raise an error
	if rsp.StatusCode != 200 {
		return "", errors.New("wolfram alpha didn't like that")
	}
	// Read response
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Alpha handles incoming requests.
// w is a ResponseWriter and r is a pointer to a request
func Alpha(w http.ResponseWriter, r *http.Request) {
	// Decode request
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		// Get text for query
		question, present := t["text"].(string)
		if present {
			// Ask wolfram the question
			answer, err := AlphaReq(question)
			if err == nil {
				// Encode response
				responseBody := map[string]interface{}{"text": answer}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(responseBody)
				if err != nil {
					// Error handling if response could not be encoded
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Could not encode response"))
					check(err)
				}
			} else {
				// Error handling if wolfram could not be contacted
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Something went wrong when contacting the Alpha API"))
				check(err)
			}
		} else {
			// Error handling if request is not in the correct JSON format
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("Incorrectly Formatted JSON"))
			check(err)
		}
	} else {
		// Error handling if request does not contain a JSON body
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("Input not in JSON format"))
		check(err)
	}
}

func main() {
	// Load keys into a map
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	defer keyFile.Close()
	byteValue, err := ioutil.ReadAll(keyFile)
	check(err)
	err = json.Unmarshal(byteValue, &keys)
	check(err)
	// Router
	r := mux.NewRouter()
	// End point
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	err = http.ListenAndServe(":3001", r)
	check(err)
}
