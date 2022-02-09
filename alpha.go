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

func AlphaReq(question string) (string, error) {
	keyFile, err := os.Open("keys.json")
	if err != nil {
		fmt.Println(err)
	}
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	var keys Keys
	json.Unmarshal(byteValue, &keys)
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
		answer, err := AlphaReq(t["text"].(string))
		responseBody := map[string]interface{}{"text": answer}
		w.WriteHeader(http.StatusOK)
		fmt.Println(err)
		json.NewEncoder(w).Encode(responseBody)
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	http.ListenAndServe(":3000", r)
}
