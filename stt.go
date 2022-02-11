package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&t)
	speech := t["speech"].(string)
	data, error := base64.StdEncoding.DecodeString(speech)
	check(error)
	text, err := handleResponse(data)
	check(err)
	u := map[string]interface{}{"text": text}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func handleResponse(speech []byte) (string, error) {
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	var keys Keys
	json.Unmarshal(byteValue, &keys)
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	check(err)
	req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	rsp, err := client.Do(req)
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	t := map[string]interface{}{}
	json.NewDecoder(bytes.NewReader(body)).Decode(&t)
	text := t["DisplayText"].(string)
	return text, err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", handleRequest).Methods("POST")
	http.ListenAndServe(":3002", r)
}
