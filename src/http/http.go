package http

import (
	"../g"
	//"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func ValidIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
}

func StartHttp() {

	configApiRoutes()
	configIndexRoutes()
	seelog.Info("[func:StartHttp] starting to listen on ", g.Cfg.Port)
	s := fmt.Sprintf(":%d", g.Cfg.Port)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		log.Fatalln("[StartHttp]", err)
	}
	os.Exit(0)
}
