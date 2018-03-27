package httpServer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"project-42/common"
	_ "project-42/fsApi"
	dbDriver "project-42/mongoDriver"
)

var mangaTitlesUrl string = fmt.Sprintf("http://%v:%v/images/mangaTitles/", common.Config.Ftp.Host, common.Config.Ftp.Port)

type mainData struct {
	Manga  []dbDriver.Manga
	Ranobe []dbDriver.Ranobe
}

type readMangaData struct {
	dbDriver.Manga
	Images         []string
	CurrentChapter int
}

func processSingleMangaTitles(manga dbDriver.Manga) dbDriver.Manga {
	if len(manga.Titles) == 0 {
		manga.Titles = append(manga.Titles, "/static/mangaNoTitleImage.png")
	} else {
		for iTitle, title := range manga.Titles {
			manga.Titles[iTitle] = fmt.Sprintf("%s%s/%s", mangaTitlesUrl, manga.Url, title)
		}
	}

	return manga
}

func processMangaTitles(mangaSlice []dbDriver.Manga) []dbDriver.Manga {
	for iManga, manga := range mangaSlice {
		if len(manga.Titles) == 0 {
			mangaSlice[iManga].Titles = append(manga.Titles, "/static/mangaNoTitleImage.png")
		} else {
			for iTitle, title := range manga.Titles {
				mangaSlice[iManga].Titles[iTitle] = fmt.Sprintf("%s%s/%s", mangaTitlesUrl, manga.Url, title)
			}
		}
	}

	return mangaSlice
}

func httpMain(w http.ResponseWriter, r *http.Request) {
	data := mainData{
		Manga:  processMangaTitles(dbDriver.GetMangaAll()),
		Ranobe: dbDriver.GetRanobeAll(),
	}

	t, err := template.ParseFiles("./HTML/main.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/main.gohtml")
		http.Error(w, err.Error(), 500)
		return
	}
	t.Execute(w, data)
}

func httpMangaMain(w http.ResponseWriter, r *http.Request) {
	data := processMangaTitles(dbDriver.GetMangaAll())

	t, err := template.ParseFiles("./HTML/mangaMain.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaMain.gohtml")
		http.Error(w, err.Error(), 500)
		return
	}
	t.Execute(w, data)
}

func httpMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data = processSingleMangaTitles(data)

	t, err := template.ParseFiles("./HTML/mangaInfo.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaInfo.gohtml")
		http.Error(w, err.Error(), 500)
		return
	}
	t.Execute(w, data)
}

func httpMangaRead(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	manga, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data := readMangaData{
		Manga:          processSingleMangaTitles(manga),
		Images:         dbDriver.GetMangaImages(mux.Vars(r)["name"], chapNumber),
		CurrentChapter: chapNumber,
	}

	t, err := template.ParseFiles("./HTML/mangaRead.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaRead.gohtml")
		http.Error(w, err.Error(), 500)
		return
	}
	t.Execute(w, data)
}

func apiMain(w http.ResponseWriter, r *http.Request) {
	data := mainData{
		Manga:  processMangaTitles(dbDriver.GetMangaAll()),
		Ranobe: dbDriver.GetRanobeAll(),
	}
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiMain"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiMangaMain(w http.ResponseWriter, r *http.Request) {
	data := processMangaTitles(dbDriver.GetMangaAll())
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiMangaMain"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data = processSingleMangaTitles(data)
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiMangaInfo"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiMangaRead(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	manga, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data := readMangaData{
		Manga:          processSingleMangaTitles(manga),
		Images:         dbDriver.GetMangaImages(mux.Vars(r)["name"], chapNumber),
		CurrentChapter: chapNumber,
	}
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiMangaRead"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}
