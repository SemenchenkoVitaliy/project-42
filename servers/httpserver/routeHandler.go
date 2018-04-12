package httpserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

var templates = template.Must(template.ParseGlob("./HTML/*.gohtml"))

func writeServerInternalError(w http.ResponseWriter, err error, text string) {
	common.CreateLog(err, text)
	http.Error(w, err.Error(), 500)
}

func processMangaTitles(manga dbDriver.Manga) dbDriver.Manga {
	if len(manga.Titles) == 0 {
		manga.Titles = append(manga.Titles, "/static/mangaNoTitleImage.png")
	} else {
		for iTitle, title := range manga.Titles {
			manga.Titles[iTitle] = fmt.Sprintf(
				"http://img.%v/images/mangaTitles/%v/%v",
				common.Config.PublicUrl,
				manga.Url,
				title,
			)
		}
	}

	return manga
}

func httpAdmin(w http.ResponseWriter, r *http.Request) {
	manga, err := dbDriver.GetMangaAllMin()
	if err != nil {
		writeServerInternalError(w, err, "Get minimized manga list")
		return
	}

	data := struct {
		Manga     []dbDriver.MangaMin
		PublicUrl string
	}{
		Manga:     manga,
		PublicUrl: common.Config.PublicUrl,
	}

	err = templates.ExecuteTemplate(w, "admin", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template admin")
		return
	}
}

func httpAdminManga(w http.ResponseWriter, r *http.Request) {
	manga, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}

	data := struct {
		dbDriver.Manga
		PublicUrl string
	}{
		Manga:     manga,
		PublicUrl: common.Config.PublicUrl,
	}

	err = templates.ExecuteTemplate(w, "adminManga", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template adminManga")
		return
	}
}

func httpMain(w http.ResponseWriter, r *http.Request) {
	manga, err := dbDriver.GetMangaAll()
	if err != nil {
		writeServerInternalError(w, err, "Get manga all")
		return
	}

	ranobe, err := dbDriver.GetRanobeAll()
	if err != nil {
		writeServerInternalError(w, err, "Get ranobe all")
		return
	}

	for index, item := range manga {
		manga[index] = processMangaTitles(item)
	}

	data := struct {
		Manga  []dbDriver.Manga
		Ranobe []dbDriver.Ranobe
	}{
		Manga:  manga,
		Ranobe: ranobe,
	}

	err = templates.ExecuteTemplate(w, "main", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template main")
		return
	}
}

func httpMangaMain(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetMangaAll()
	if err != nil {
		writeServerInternalError(w, err, "Get manga all")
		return
	}

	for index, item := range data {
		data[index] = processMangaTitles(item)
	}

	err = templates.ExecuteTemplate(w, "mangaMain", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template mangaMain")
		return
	}
}

func httpMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}

	data = processMangaTitles(data)

	err = templates.ExecuteTemplate(w, "mangaInfo", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template mangaInfo")
		return
	}
}

func httpMangaRead(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	manga, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}
	images, err := dbDriver.GetMangaImages(mux.Vars(r)["name"], chapNumber)
	if err != nil {
		writeServerInternalError(w, err, "Manga images database request of "+mux.Vars(r)["name"]+"-"+mux.Vars(r)["chapter"])
		return
	}

	data := struct {
		dbDriver.Manga
		Images         []string
		CurrentChapter int
		PublicUrl      string
	}{
		Manga:          manga,
		Images:         images,
		CurrentChapter: chapNumber,
		PublicUrl:      common.Config.PublicUrl,
	}

	err = templates.ExecuteTemplate(w, "mangaRead", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template mangaRead")
		return
	}
}
