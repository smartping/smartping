package http

import (
	"net/http"
	"log"
	"database/sql"
	"../g"
)

func configIndexRoutes(port int, state *g.State ,db *sql.DB ,config g.Config){
	//Index
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state.Lock.Lock()
		defer state.Lock.Unlock()
		r.ParseForm()
		state.Showtype="out"
		if len(r.Form["t"]) > 0 {
			state.Showtype = r.Form["t"][0]
		}
		state.Conf = config
		//log.Println(state)
		err := index.Execute(w, state)
		if err != nil {
			log.Println("ERR:",err)
		}
	})

}
