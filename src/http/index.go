package http

import (
	"github.com/smartping/smartping/src/g"
	"net/http"
	"path/filepath"
	"strings"
)

func configIndexRoutes() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/") {
			if !g.IsExist(filepath.Join(g.Root, "/html", r.URL.Path, "index.html")) {
				http.NotFound(w, r)
				return
			}
		}
		http.FileServer(http.Dir(filepath.Join(g.Root, "/html"))).ServeHTTP(w, r)
	})

}
