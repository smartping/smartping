package http

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"encoding/json"
	//"strconv"
	"database/sql"
	//"time"
	"../funcs"
	"../g"
	//"github.com/go-resty/resty"
	//"bytes"
	"os"
)

// Init of the Web Page template.
var index = template.Must(template.New("index.tpl").Delims("<%", "%>").Funcs(template.FuncMap{"compare": funcs.Compare,"timestr":funcs.Timestr,"md5str":funcs.Md5str}).Parse(indexTemplate))
var topology = template.Must(template.New("topology.tpl").Delims("<%", "%>").Funcs(template.FuncMap{"json": json.Marshal}).Parse(topologyTemplate))


func StartHttp(port int, state *g.State ,db *sql.DB ,config g.Config) {

	configApiRoutes(port,state,db,config)
	configIndexRoutes(port,state,db,config)
	configTopologyRoutes(port,state,db,config)

	s := fmt.Sprintf(":%d", port)
	log.Println("starting to listen on ", s)
	log.Printf("Get status on http://localhost%s/status", s)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		log.Println("ERR:",err)
	}
	log.Println("Server on 8899 stopped")
	os.Exit(0)
}
