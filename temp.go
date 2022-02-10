package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type WordPopularity struct {
	Word       string
	Popularity float64
}

type AllWords struct {
	words []WordPopularity
}

func main() {
	jsonFile, err := os.Open("word_freq.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened word_freq.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}
	var result map[string]float64
	json.Unmarshal(byteValue, &result)
	for key, val := range result {
		fmt.Println(key, val)

	}
	fmt.Println("Successfully Opened users.json")
}
