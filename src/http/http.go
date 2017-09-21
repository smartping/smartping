package http

import (
	"../g"
	"database/sql"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"net/http"
	"os"
)

func StartHttp(db *sql.DB, config *g.Config) {

	configApiRoutes(db, config)
	configIndexRoutes()
	seelog.Debug("[func:StartHttp] starting to listen on", config.Port)
	s := fmt.Sprintf(":%d", config.Port)
	//log.Println("starting to listen on ", s)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		seelog.Error("[StartHttp] ", err)
	}
	os.Exit(0)
}
