package main

import (
	"encoding/json"
	"log"
	"os"
)

// Opening (or creating) config file in JSON format
func readConfig(filename string) Config {
	config := Config{}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatal("Config Not Found!")
	} else {
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	}
	return config
}
