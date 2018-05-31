package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/dbDriver/mongo"
	"github.com/SemenchenkoVitaliy/project-42/mangaLoader"
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
	"github.com/gorilla/mux"
)

var (
	db      = dbDriver.NewDatabase()
	dbUsers = dbDriver.NewDatabaseUsers()

	fsPublicUrl string
	publicUrl   string
)

func checkUser(w http.ResponseWriter, r *http.Request) (user dbDriver.User, err error) {
	c, err := r.Cookie("session")
	if err != nil {
		return user, err
	}
	return dbUsers.FindUser(c.Value)
}

func indexGET(w http.ResponseWriter, r *http.Request) {
	mangaUrls, err := db.GetMangaUrls(0, 10, "upddate")
	if err != nil {
		netutils.InternalError(w, err, "Get top manga urls")
		return
	}

	manga, err := db.GetMangaMultiple(mangaUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top manga")
		return
	}

	for index, item := range manga {
		manga[index].Covers = utils.ProcessCovers(item.Covers, item.Url)
	}

	ranobeUrls, err := db.GetRanobeUrls(0, 10, "upddate")
	if err != nil {
		netutils.InternalError(w, err, "Get top ranobe urls")
		return
	}

	ranobe, err := db.GetRanobeMultiple(ranobeUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top ranobe")
		return
	}

	data := struct {
		Manga  []dbDriver.Product
		Ranobe []dbDriver.Product
	}{
		Manga:  manga,
		Ranobe: ranobe,
	}

	result, err := json.Marshal(data)
	if err != nil {
		netutils.InternalError(w, err, "JSON convert in roorGET")
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(result))
}

func mangaGET(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	mangaUrls, err := db.GetMangaUrls(page*10, 10, "name")
	if err != nil {
		netutils.InternalError(w, err, "Get top manga urls")
		return
	}

	data, err := db.GetMangaMultiple(mangaUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top manga")
		return
	}

	for index, item := range data {
		data[index].Covers = utils.ProcessCovers(item.Covers, item.Url)
	}

	result, err := json.Marshal(data)
	if err != nil {
		netutils.InternalError(w, err, "JSON convert in mangaAllGET")
		return
	}
	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(result))
}

func mangaInfoGET(w http.ResponseWriter, r *http.Request) {
	data, err := db.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		utils.Log(err, "Get manga info "+mux.Vars(r)["name"])

	}

	data.Covers = utils.ProcessCovers(data.Covers, data.Url)

	result, err := json.Marshal(data)
	if err != nil {
		netutils.InternalError(w, err, "JSON convert in mangaOneGET"+mux.Vars(r)["name"])
		return
	}
	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(result))
}

func mangaChapterGET(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	data, err := db.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		utils.Log(err, "GetMangaSingle "+mux.Vars(r)["name"])
		netutils.InternalError(w, err, "Get top 10 manga")

	}

	images := utils.ProcessPages(
		data.Chapters[chapNumber].Pages,
		mux.Vars(r)["name"],
		mux.Vars(r)["chapter"])

	result, err := json.Marshal(images)
	if err != nil {
		netutils.InternalError(w, err, "JSON convert in mangaChapterGET"+mux.Vars(r)["name"])
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(result))
}

func indexPOST(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "reloadHtml":
		mainServer.UpdateHtml()
	default:
		http.Error(w, "no such action: "+action, 400)
		return
	}
	w.WriteHeader(200)
}

func mangaPOST(w http.ResponseWriter, r *http.Request) {
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
		url := strings.ToLower(strings.Replace(name, " ", "_", -1))

		manga := dbDriver.Product{
			Name:     name,
			Url:      url,
			Size:     0,
			SrcUrl:   "",
			AddDate:  time.Now(),
			UpdDate:  time.Now(),
			Covers:   []string{},
			Chapters: []dbDriver.Chapter{},
		}

		err := db.AddManga(manga)
		if err != nil {
			http.Error(w, "no such action", 400)
		}

	default:
		http.Error(w, "no such action: "+action, 400)
		return
	}

	w.WriteHeader(200)
}

func mangaInfoPOST(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")

	switch action {
	case "update":
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
		go mangaLoader.UpdateManga(mux.Vars(r)["name"])
	case "remove":
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
		db.RemoveManga(mux.Vars(r)["name"])
	case "changeName":
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
		db.SetMangaName(mux.Vars(r)["name"], r.FormValue("name"))
	case "addCover":
		r.ParseMultipartForm(32 << 20)
		fhs := r.MultipartForm.File["file"]
		for _, header := range fhs {
			if header.Filename == "" {
				return
			}
			file, err := header.Open()
			defer file.Close()

			if err != nil {
				utils.Log(err, "form file in addCover in"+mux.Vars(r)["name"])
				http.Error(w, err.Error(), 500)
			}

			var buf bytes.Buffer
			io.Copy(&buf, file)

			filePath := fmt.Sprintf(
				"/images/mangaCovers/%v/%v",
				mux.Vars(r)["name"],
				header.Filename,
			)
			mainServer.WriteFile(filePath, buf.Bytes())

			db.AddMangaCover(mux.Vars(r)["name"], header.Filename)
		}
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	case "remCover":
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
		db.RemoveMangaCover(mux.Vars(r)["name"], r.FormValue("fileName"))
	case "addChapter":
		db.AddMangaChapterEmpty(mux.Vars(r)["name"], r.FormValue("name"))
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	default:
		http.Error(w, "no such action: "+action, 400)
		return
	}
	w.WriteHeader(200)
}

func mangaChapterPOST(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, "Incorrect request", 400)
	}

	action := r.FormValue("action")
	switch action {
	case "changeName":
		name := r.FormValue("name")

		db.SetMangaChapterName(mux.Vars(r)["name"], chapNumber, name)
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	case "remove":
		db.RemoveMangaChapter(mux.Vars(r)["name"], chapNumber)
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	case "addPages":
		r.ParseMultipartForm(32 << 20)
		fhs := r.MultipartForm.File["file"]
		pages := []string{}
		for _, header := range fhs {
			if header.Filename == "" {
				return
			}
			file, err := header.Open()
			defer file.Close()

			if err != nil {
				utils.Log(err, "form file in addPages in"+mux.Vars(r)["name"])
				http.Error(w, err.Error(), 500)
			}

			var buf bytes.Buffer
			io.Copy(&buf, file)

			filePath := fmt.Sprintf(
				"/images/manga/%v/%v/%v",
				mux.Vars(r)["name"],
				chapNumber,
				header.Filename,
			)
			mainServer.WriteFile(filePath, buf.Bytes())
			pages = append(pages, header.Filename)
		}

		db.AddMangaChapterPages(mux.Vars(r)["name"], chapNumber, pages)
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	case "removePages":
		db.RemoveMangaChapterPages(mux.Vars(r)["name"], chapNumber)
		mainServer.UpdateProductCache("manga", mux.Vars(r)["name"])
	default:
		http.Error(w, "no such action: "+action, 400)
		return
	}

	w.WriteHeader(200)
}
