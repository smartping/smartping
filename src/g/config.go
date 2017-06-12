package g

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

//var DataLock = new
var DLock sync.Mutex

// Opening (or creating) config file in JSON format
func ReadConfig(filename string) Config {
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
