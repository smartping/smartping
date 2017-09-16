package http

import (
	"../g"
	"github.com/gy-games-libs/file"
	"net/http"
	"path/filepath"
	"strings"
)

func configIndexRoutes() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			if !file.IsExist(filepath.Join(g.Root, "/html", r.URL.Path, "index.html")) {
				http.NotFound(w, r)
				return
			}
		}
		http.FileServer(http.Dir(filepath.Join(g.Root, "/html"))).ServeHTTP(w, r)
	})

}
