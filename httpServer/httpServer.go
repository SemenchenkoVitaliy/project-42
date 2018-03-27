package httpServer

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"project-42/common"
)

func Start() {
	r := mux.NewRouter()

	r.HandleFunc("/", httpMain).Methods("GET")
	r.HandleFunc("/manga", httpMangaMain).Methods("GET")
	r.HandleFunc("/manga/{name}", httpMangaInfo).Methods("GET")
	r.HandleFunc("/manga/{name}/{chapter}", httpMangaRead).Methods("GET")

	r.HandleFunc("/api", apiMain).Methods("GET")
	r.HandleFunc("/api/manga", apiMangaMain).Methods("GET")
	r.HandleFunc("/api/manga/{name}", apiMangaInfo).Methods("GET")
	r.HandleFunc("/api/manga/{name}/{chapter}", apiMangaRead).Methods("GET")

	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("http server is being opened on %v:%v\n", common.Config.Http.Host, common.Config.Http.Port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.Http.Host, common.Config.Http.Port), nil)
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.Http.Host, common.Config.Http.Port))
	}
}
