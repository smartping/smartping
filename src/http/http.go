package http

import (
	"../g"
	"database/sql"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"log"
	"net/http"
	"os"
)

func StartHttp(db *sql.DB, config *g.Config) {

	configApiRoutes(db, config)
	configIndexRoutes()
	seelog.Info("[func:StartHttp] starting to listen on ", config.Port)
	s := fmt.Sprintf(":%d", config.Port)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		log.Fatalln("[StartHttp]", err)
	}
	os.Exit(0)
}
