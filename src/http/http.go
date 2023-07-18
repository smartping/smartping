package http

import (
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func ValidDomain(domainName string) bool {
	domainName = strings.Trim(domainName, " ")
	domainre, _ := regexp.Compile(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z]{0,62})\.?`)
	if domainre.MatchString(domainName) {
		return true
	}
	return false
}

func ValidDomaina(domainName string) bool {
	domainName = strings.Trim(domainName, " ")
	domainre, _ := regexp.Compile(`[a-zA-Z]{0,62}-.*`)
	if domainre.MatchString(domainName) {
		return true
	}
	return false
}

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

func AuthUserIp(RemoteAddr string) bool {
	if len(g.AuthUserIpMap) == 0 {
		return true
	}
	ips := strings.Split(RemoteAddr, ":")
	if len(ips) == 2 {
		if _, ok := g.AuthUserIpMap[ips[0]]; ok {
			return true
		}
	}
	return false
}

func AuthAgentIp(RemoteAddr string, drt bool) bool {
	if drt {
		if len(g.AuthUserIpMap) == 0 {
			return true
		}
	}
	if len(g.AuthAgentIpMap) == 0 {
		return true
	}
	ips := strings.Split(RemoteAddr, ":")
	if len(ips) == 2 {
		if _, ok := g.AuthAgentIpMap[ips[0]]; ok {
			return true
		}
	}
	return false
}

func GraphText(x int, y int, txt string) chart.Renderer {
	f, _ := chart.GetDefaultFont()
	rhart, _ := chart.PNG(300, 130)
	chart.Draw.Text(rhart, txt, x, y, chart.Style{
		FontColor: drawing.ColorBlack,
		FontSize:  10,
		Font:      f,
	})
	return rhart
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
