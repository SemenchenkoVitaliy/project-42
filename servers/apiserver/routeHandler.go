package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/mangaLoader"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

type mainData struct {
	Manga  []dbDriver.Manga
	Ranobe []dbDriver.Ranobe
}

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
				"common.Config.PublicUrl%v/%v",
				common.Config.PublicUrl,
				manga.Url,
				title,
			)
		}
	}

	return manga
}

func apiGetMain(w http.ResponseWriter, r *http.Request) {
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

	data := mainData{
		Manga:  manga,
		Ranobe: ranobe,
	}

	stringifiedData, err := json.Marshal(data)
	if err != nil {
		writeServerInternalError(w, err, "JSON convert in apiGetMain")
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaMain(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetMangaAllMin()
	if err != nil {
		writeServerInternalError(w, err, "Get minimized manga all")
		return
	}

	stringifiedData, err := json.Marshal(data)
	if err != nil {
		writeServerInternalError(w, err, "JSON convert in apiGetMangaMain")
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := dbDriver.GetManga(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}

	data = processMangaTitles(data)
	stringifiedData, err := json.Marshal(data)
	if err != nil {
		writeServerInternalError(w, err, "JSON convert in apiGetMangaInfo"+mux.Vars(r)["name"])
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

	data, err := dbDriver.GetMangaImages(mux.Vars(r)["name"], chapNumber)
	if err != nil {
		writeServerInternalError(w, err, fmt.Sprintf(
			"Manga images database request of %v-%v",
			mux.Vars(r)["name"],
			mux.Vars(r)["chapter"],
		))
		return
	}

	for index, image := range data {
		data[index] = fmt.Sprintf(
			"http://img.%v/images/manga/%v/%v/%v",
			common.Config.PublicUrl,
			mux.Vars(r)["name"],
			mux.Vars(r)["chapter"],
			image,
		)
	}

	stringifiedData, err := json.Marshal(data)
	if err != nil {
		writeServerInternalError(w, err, "JSON convert in apiGetMangaRead"+mux.Vars(r)["name"])
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

		MkDir("images/manga/" + url)
		MkDir("images/mangaTitles/" + url)

		mangaLoader.AddManga(url)
	default:
		http.Error(w, "no such action: "+action, 400)
		return
	}

	apiGetMangaMain(w, r)
}

func apiChangeMangaInfo(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "update":
		apiGetMangaMain(w, r)
		mangaLoader.UpdateManga(mux.Vars(r)["name"])
	case "remove":
		dbDriver.RemoveManga(mux.Vars(r)["name"])
		apiGetMangaMain(w, r)
	case "changeName":
		name := r.FormValue("name")
		dbDriver.ChangeMangaName(mux.Vars(r)["name"], name)
		apiGetMangaInfo(w, r)
	case "addTitle":
		file, header, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			common.CreateLog(err, "form file in addTitle in"+mux.Vars(r)["name"])
			http.Error(w, err.Error(), 500)
		}

		var buf bytes.Buffer
		io.Copy(&buf, file)

		filePath := fmt.Sprintf(
			"images/mangaTitles/%v/%v",
			mux.Vars(r)["name"],
			header.Filename,
		)
		WriteFile(filePath, buf.Bytes())
		dbDriver.AddMangaTitle(mux.Vars(r)["name"], header.Filename)
		apiGetMangaInfo(w, r)
	case "remTitle":
		fileName := r.FormValue("fileName")
		dbDriver.RemoveMangaTitle(mux.Vars(r)["name"], fileName)
		apiGetMangaInfo(w, r)
	default:
		http.Error(w, "no such action", 400)
		return
	}
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

	apiGetMangaInfo(w, r)
}
