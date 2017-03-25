package main

import (
	"flag"
	"log"
	"database/sql"
	"sync"
)
// Init config
var filename = flag.String("f", "config.json", "JSON configuration file")
var httpPort = flag.Int("p", 8899, "HTTP port")
var lock sync.Mutex

var Version = "0.1.3"
// Main function
func main() {
	flag.Parse()
	// Config
	var config Config
	log.Println("Opening config file: ", *filename)
	config = readConfig(*filename)
	config.Ver = Version
	log.Printf("Config loaded")
	log.Println("LogDB : "+config.Db)
	db, err := sql.Open("sqlite3", config.Db)
	if err != nil {
		log.Print(err)
	}
	// Running
	res := make(chan TargetStatus)
	state := NewState()
	state.Localname = config.Name
	state.Localip = config.Ip
	for _, target := range config.Targets {
		startPing(db, config, target, res)
	}
	// HTTP
	go startHttp(*httpPort, state, db , config)
	for {
		status := <-res
		state.Lock.Lock()
		state.State[status.Target] = status
		state.Lock.Unlock()
	}

}
