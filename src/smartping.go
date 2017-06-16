package main

import (
	"flag"
	"log"
	"database/sql"
	"./g"
	"./funcs"
	"./http"
	"os"
)
// Init config
var filename = flag.String("f", "config.json", "JSON configuration file")
var httpPort = flag.Int("p", 8899, "HTTP port")
//var lock sync.Mutex

var Version = "0.2.3"
// Main function
func main() {
	flag.Parse()
	// Config
	var config g.Config
	log.Println("Opening config file: ", *filename)
	config = g.ReadConfig(*filename)
	if config.Name==""{
		config.Name,_ = os.Hostname()
	}
	if config.Ip==""{
		config.Ip = "127.0.0.1"
	}
	config.Ver = Version
	config.Db = funcs.GetRoot()+"/db/database.db"
	log.Printf("Config loaded")
	//log.Println("LogDB : "+config.Db)
	db, err := sql.Open("sqlite3", config.Db)
	if err != nil {
		log.Print(err)
	}
	// Running
	res := make(chan g.TargetStatus)
	state := funcs.NewState()
	state.Localname = config.Name
	state.Localip = config.Ip
	for _, target := range config.Targets {
		funcs.StartPing(db, config, target, res)
	}
	// HTTP
	go http.StartHttp(*httpPort, state, db , config)
	for {
		status := <-res
		state.Lock.Lock()
		state.State[status.Target] = status
		state.Lock.Unlock()
	}
}
