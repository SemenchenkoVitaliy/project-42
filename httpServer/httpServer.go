package httpServer

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

func Start() {

	r := mux.NewRouter()

	r.HandleFunc("/", httpMain).Methods("GET")
	r.HandleFunc("/manga", httpMangaMain).Methods("GET")
	r.HandleFunc("/manga/{name}", httpMangaInfo).Methods("GET")
	r.HandleFunc("/manga/{name}/{chapter}", httpMangaRead).Methods("GET")

	r.HandleFunc("/api", apiGetMain).Methods("GET")
	r.HandleFunc("/api/manga", apiGetMangaMain).Methods("GET")
	r.HandleFunc("/api/manga/{name}", apiGetMangaInfo).Methods("GET")
	r.HandleFunc("/api/manga/{name}/{chapter}", apiGetMangaRead).Methods("GET")

	r.HandleFunc("/api/manga", apiChangeMangaMain).Methods("POST")
	r.HandleFunc("/api/manga/{name}", apiChangeMangaInfo).Methods("POST")
	r.HandleFunc("/api/manga/{name}/{chapter}", apiChangeMangaChapter).Methods("POST")

	r.HandleFunc("/admin", httpAdmin).Methods("GET")
	r.HandleFunc("/admin/manga/{name}", httpAdminManga).Methods("GET")

	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("http server is being opened on %v:%v\n", common.Config.Http.Host, common.Config.Http.Port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.Http.Host, common.Config.Http.Port), nil)
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.Http.Host, common.Config.Http.Port))
	}
}
