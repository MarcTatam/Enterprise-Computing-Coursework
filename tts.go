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

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		text, present := t["text"].(string)
		if present {
			speech, err := handleResponse(xmlFormat(text))
			if err == nil {
				u := map[string]interface{}{"speech": speech}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Could not encode JSON"))
					check(err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Something went wrong contacting the text to speech api"))
				check(err)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("Incorrect JSON format"))
			check(err)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Request not in JSON format"))
		check(err)
	}
}

func xmlFormat(text string) string {
	return "<?xml version=\"1.0\"?><speak version=\"1.0\" xml:lang=\"en-US\"><voice xml:lang=\"en-US\" name=\"en-US-JennyNeural\">" + text + "</voice></speak>"
}

func handleResponse(text string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewBuffer([]byte(text)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		fmt.Println(rsp.StatusCode)
		return nil, errors.New("microsoft didn't like that")
	}
	return b, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	check(err)
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	err = json.Unmarshal(byteValue, &keys)
	check(err)
	r := mux.NewRouter()
	// document
	r.HandleFunc("/tts", handleRequest).Methods("POST")
	err = http.ListenAndServe(":3003", r)
	check(err)
}
