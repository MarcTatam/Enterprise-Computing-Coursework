package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Keys struct {
	Speech string `json:"Speech"`
	Alpha  string `json:"Alpha"`
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
	var keys Keys
	json.Unmarshal(byteValue, &keys)
	fmt.Println(keys.Alpha)
	check(err)
}
