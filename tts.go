package main

import (
	"bytes"
	"encoding/json"
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

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&t)
	text := t["text"].(string)
	speech, err := handleResponse(xmlFormat(text))
	check(err)
	u := map[string]interface{}{"speech": speech}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func xmlFormat(text string) string {
	return "<?xml version=\"1.0\"?><speak version=\"1.0\" xml:lang=\"en-US\"><voice xml:lang=\"en-US\" name=\"en-US-JennyNeural\">" + text + "</voice></speak>"
}

func handleResponse(text string) ([]byte, error) {
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	check(err)
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	var keys Keys
	json.Unmarshal(byteValue, &keys)
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewBuffer([]byte(text)))
	check(err)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", keys.Speech)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	rsp, err2 := client.Do(req)
	check(err2)
	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	check(err)
	fmt.Println(rsp.StatusCode)
	return b, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/tts", handleRequest).Methods("POST")
	http.ListenAndServe(":3003", r)
}
