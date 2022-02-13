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

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err == nil {
		speech := t["speech"].(string)
		data, err := base64.StdEncoding.DecodeString(speech)
		if err == nil {
			text, err := handleResponse(data)
			if err == nil {
				check(err)
				u := map[string]interface{}{"text": text}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Something went wrong encoding the response"))
					check(err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Something went wrong contacting the STT provider"))
				check(err)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("Speech is not a Base 64 encoded string"))
			check(err)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Request format is not JSON"))
		check(err)
	}

}

func handleResponse(speech []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	check(err)
	req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	rsp, err := client.Do(req)
	if rsp.StatusCode != 200 {
		return "", errors.New("microsoft didn't like that")
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	t := map[string]interface{}{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&t)
	check(err)
	text := t["DisplayText"].(string)
	return text, err
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
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	err = json.Unmarshal(byteValue, &keys)
	check(err)
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", handleRequest).Methods("POST")
	err = http.ListenAndServe(":3002", r)
	check(err)
}
