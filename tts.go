package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Keys struct {
	Speech string `json:"Speech"`
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
	fmt.Println("Successfully Opened users.json")
	defer keyFile.Close()
	byteValue, _ := ioutil.ReadAll(keyFile)
	var keys Keys
	json.Unmarshal(byteValue, &keys)
	fmt.Println(keys.Speech)
	check(err)
}
