package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	DBName       string
	DBHost       string
	DBCollection string
}

func getConfigVars(fileName string) (*Configuration, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	var configuration Configuration
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)
	if err != nil {
		return nil, err
	}

	return &configuration, err
}
