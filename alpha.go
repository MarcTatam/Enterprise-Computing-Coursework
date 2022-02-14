package main

import (
	"encoding/json"
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

func AlphaReq(question string) (string, error) {
	client := &http.Client{}
	uri := URI + "?appid=" + keys.Alpha + "&i=" + url.QueryEscape(question)
	req, err := http.NewRequest("GET", uri, nil)
	rsp, err := client.Do(req)
	fmt.Println(rsp.StatusCode)
	b, err := io.ReadAll(rsp.Body)
	return string(b), nil
}

func Alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		question := t["text"].(string)
		answer, err := AlphaReq(question)
		if err == nil {
			responseBody := map[string]interface{}{"text": answer}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(responseBody)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong when contacting the Alpha API"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("Input not in JSON format"))
		check(err)
	}
}

func main() {
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	defer keyFile.Close()
	byteValue, err := ioutil.ReadAll(keyFile)
	check(err)
	err = json.Unmarshal(byteValue, &keys)
	check(err)
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	err = http.ListenAndServe(":3000", r)
	check(err)
}
