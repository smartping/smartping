package http

import (
	"../g"
	"database/sql"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"net/http"
	"os"
	"time"
)

func StartHttp(db *sql.DB, config *g.Config) {

	configApiRoutes(db, config)
	configIndexRoutes()
	seelog.Info("[func:StartHttp] starting to listen on ", config.Port)
	s := fmt.Sprintf(":%d", config.Port)
	//log.Println("starting to listen on ", s)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		seelog.Error("[StartHttp] ", err)
		time.Sleep(3*time.Second)
	}
	os.Exit(0)
}
