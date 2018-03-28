package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/fsApi"
	"github.com/SemenchenkoVitaliy/project-42/mangaLoader"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
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

type htmlTemplates struct {
	Main      *template.Template
	MangaMain *template.Template
	MangaInfo *template.Template
	MangaRead *template.Template
}

var templates htmlTemplates = htmlTemplates{}

func init() {
	var err error
	templates.Main, err = template.ParseFiles("./HTML/main.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/main.gohtml")
		return
	}
	templates.MangaMain, err = template.ParseFiles("./HTML/mangaMain.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaMain.gohtml")
		return
	}
	templates.MangaInfo, err = template.ParseFiles("./HTML/mangaInfo.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaInfo.gohtml")
		return
	}
	templates.MangaRead, err = template.ParseFiles("./HTML/mangaRead.gohtml")
	if err != nil {
		common.CreateLog(err, "Parse ./HTML/mangaRead.gohtml")
		return
	}
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

	err := templates.Main.Execute(w, data)
	if err != nil {
		common.CreateLog(err, "Execute template main")
		http.Error(w, err.Error(), 500)
		return
	}
}

func httpMangaMain(w http.ResponseWriter, r *http.Request) {
	data := processMangaTitles(dbDriver.GetMangaAll())

	err := templates.MangaMain.Execute(w, data)
	if err != nil {
		common.CreateLog(err, "Execute template mangaMain")
		http.Error(w, err.Error(), 500)
		return
	}
}

func httpMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data = processSingleMangaTitles(data)

	err = templates.MangaInfo.Execute(w, data)
	if err != nil {
		common.CreateLog(err, "Execute template mangaInfo")
		http.Error(w, err.Error(), 500)
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
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data := readMangaData{
		Manga:          processSingleMangaTitles(manga),
		Images:         dbDriver.GetMangaImages(mux.Vars(r)["name"], chapNumber),
		CurrentChapter: chapNumber,
	}

	err = templates.MangaRead.Execute(w, data)
	if err != nil {
		common.CreateLog(err, "Execute template mangaRead")
		http.Error(w, err.Error(), 500)
		return
	}
}

func apiGetMain(w http.ResponseWriter, r *http.Request) {
	data := mainData{
		Manga:  processMangaTitles(dbDriver.GetMangaAll()),
		Ranobe: dbDriver.GetRanobeAll(),
	}
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiGetMain"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaMain(w http.ResponseWriter, r *http.Request) {
	data := processMangaTitles(dbDriver.GetMangaAll())
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiGetMangaMain"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		common.CreateLog(err, "Manga database request of "+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	data = processSingleMangaTitles(data)
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		common.CreateLog(err, "JSON convert in apiGetMangaInfo"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaRead(w http.ResponseWriter, r *http.Request) {
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
		common.CreateLog(err, "JSON convert in apiGetMangaRead"+mux.Vars(r)["name"])
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiChangeMangaMain(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "add":
		url := r.FormValue("url")
		err := mangaLoader.AddManga(url)
		if err != nil {
			http.Error(w, "no such action", 400)
		}
	case "addManual":
		name := r.FormValue("name")
		url := r.FormValue("url")

		manga := dbDriver.Manga{
			Name:     name,
			Url:      url,
			Size:     0,
			SrcUrl:   "",
			AddDate:  time.Now(),
			UpdDate:  time.Now(),
			Titles:   []string{},
			Chapters: []dbDriver.MangaChapter{},
		}

		err := dbDriver.AddManga(manga)
		if err != nil {
			http.Error(w, "no such action", 400)
		}

		fsApi.MkDir("images/manga/" + url)
		fsApi.MkDir("images/mangaTitles/" + url)

		mangaLoader.AddManga(url)
	default:
		http.Error(w, "no such action", 400)
		return
	}

	apiChangeMangaInfo(w, r)
}

func apiChangeMangaInfo(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "update":
		mangaLoader.LoadManga(mux.Vars(r)["name"])
	case "remove":
		dbDriver.RemoveManga(mux.Vars(r)["name"])
	case "changeName":
		name := r.FormValue("name")
		dbDriver.ChangeMangaName(mux.Vars(r)["name"], name)
	case "addTitle":
		file, header, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			common.CreateLog(err, "form file in addTitle in"+mux.Vars(r)["name"])
			http.Error(w, err.Error(), 500)
		}

		var buf bytes.Buffer
		io.Copy(&buf, file)
		fsApi.WriteFile("images/mangaTitles/"+mux.Vars(r)["name"]+header.Filename, buf.Bytes())
		dbDriver.AddMangaTitle(mux.Vars(r)["name"], header.Filename)
	case "remTitle":
		fileName := r.FormValue("fileName")
		dbDriver.RemoveMangaTitle(mux.Vars(r)["name"], fileName)
	default:
		http.Error(w, "no such action", 400)
		return
	}

	apiChangeMangaInfo(w, r)
}

func apiChangeMangaChapter(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "changeName":
		name := r.FormValue("name")
		chapNumber, err := strconv.ParseInt(mux.Vars(r)["chapter"], 10, 0)
		if err != nil {
			http.Error(w, "Incorrect request", 400)
		}

		dbDriver.ChangeMangaChapName(mux.Vars(r)["name"], int(chapNumber), name)
	case "remove":
		chapNumber, err := strconv.ParseInt(mux.Vars(r)["chapter"], 10, 0)
		if err != nil {
			http.Error(w, "Incorrect request", 400)
		}

		dbDriver.RemoveMangaChapter(mux.Vars(r)["name"], int(chapNumber))
	default:
		http.Error(w, "no such action", 400)
		return
	}

	apiChangeMangaInfo(w, r)
}
