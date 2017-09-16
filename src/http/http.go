package http

import (
	"../g"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
)

func StartHttp(db *sql.DB, config *g.Config) {

	configApiRoutes(db, config)
	configIndexRoutes()
	s := fmt.Sprintf(":%d", config.Port)
	log.Println("starting to listen on ", s)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		log.Println("ERR:", err)
	}
	log.Println("Server stopped")
	os.Exit(0)
}
