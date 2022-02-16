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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func AlphaReq(question string) (string, error) {
	client := &http.Client{}
	uri := URI + "?appid=" + keys.Alpha + "&i=" + url.QueryEscape(question)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode != 200 {
		return "", errors.New("wolfram alpha didn't like that")
	}
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		question, present := t["text"].(string)
		if present {
			answer, err := AlphaReq(question)
			if err == nil {
				responseBody := map[string]interface{}{"text": answer}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(responseBody)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Could not encode response"))
					check(err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Something went wrong when contacting the Alpha API"))
				check(err)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("Incorrectly Formatted JSON"))
			check(err)
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
	err = http.ListenAndServe(":3001", r)
	check(err)
}
