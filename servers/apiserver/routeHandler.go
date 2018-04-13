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

func writeServerInternalError(w http.ResponseWriter, err error, text string) {
	common.CreateLog(err, text)
	http.Error(w, err.Error(), 500)
}

func processMangaTitles(manga dbDriver.Product) dbDriver.Product {
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

func apiGetMain(w http.ResponseWriter, r *http.Request) {
	mangaUrls, err := dbDriver.GetMangaUrls(0, 10, "updDate")
	if err != nil {
		writeServerInternalError(w, err, "Get top 10 manga urls")
		return
	}

	ranobeUrls, err := dbDriver.GetRanobeUrls(0, 10, "updDate")
	if err != nil {
		writeServerInternalError(w, err, "Get top 10 ranobe urls")
		return
	}

	manga := make([]dbDriver.Product, 0, 10)

	for _, mangaUrl := range mangaUrls {
		product, ok := dbDriver.MangaCache.Find(mangaUrl)
		if !ok {
			product, err = dbDriver.GetMangaSingle(mangaUrl)
			if err != nil {
				common.CreateLog(err, "GetMangaSingle "+mangaUrl)
				continue
			}
			dbDriver.MangaCache.Add(product)
		}
		manga = append(manga, product)
	}

	if len(manga) == 0 {
		writeServerInternalError(w, err, "Get top 10 manga")
		return
	}

	ranobe, err := dbDriver.GetRanobeMultiple(ranobeUrls)
	if err != nil {
		writeServerInternalError(w, err, "Get top 10 ranobe")
		return
	}

	for index, item := range manga {
		manga[index] = processMangaTitles(item)
	}

	data := struct {
		Manga  []dbDriver.Product
		Ranobe []dbDriver.Product
	}{
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	mangaUrls, err := dbDriver.GetMangaUrls(page*20, 20, "name")
	if err != nil {
		writeServerInternalError(w, err, "Get manga urls")
		return
	}

	manga := make([]dbDriver.Product, 0, 10)

	for _, mangaUrl := range mangaUrls {
		product, ok := dbDriver.MangaCache.Find(mangaUrl)
		if !ok {
			product, err = dbDriver.GetMangaSingle(mangaUrl)
			if err != nil {
				common.CreateLog(err, "GetMangaSingle "+mangaUrl)
				continue
			}
			dbDriver.MangaCache.Add(product)
		}
		manga = append(manga, product)
	}

	if len(manga) == 0 {
		writeServerInternalError(w, err, "Get top 10 manga")
		return
	}

	stringifiedData, err := json.Marshal(manga)
	if err != nil {
		writeServerInternalError(w, err, "JSON convert in apiGetMangaMain")
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write([]byte(stringifiedData))
}

func apiGetMangaInfo(w http.ResponseWriter, r *http.Request) {
	data, ok := dbDriver.MangaCache.Find(mux.Vars(r)["name"])
	var err error
	if !ok {
		data, err = dbDriver.GetMangaSingle(mux.Vars(r)["name"])
		if err != nil {
			common.CreateLog(err, "GetMangaSingle "+mux.Vars(r)["name"])
			writeServerInternalError(w, err, "Get top 10 manga")

		}
		dbDriver.MangaCache.Add(data)
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

	images, ok := dbDriver.MangaPagesCache.Find(mux.Vars(r)["name"], chapNumber)
	if !ok {
		images, err = dbDriver.GetMangaChapterPages(mux.Vars(r)["name"], chapNumber)
		if err != nil {
			writeServerInternalError(w, err, fmt.Sprintf(
				"Manga images database request of %v-%v",
				mux.Vars(r)["name"],
				mux.Vars(r)["chapter"],
			))
			return
		}
		dbDriver.MangaPagesCache.Add(mux.Vars(r)["name"], chapNumber, images)
	}

	for index, image := range images {
		images[index] = fmt.Sprintf(
			"http://img.%v/images/manga/%v/%v/%v",
			common.Config.PublicUrl,
			mux.Vars(r)["name"],
			mux.Vars(r)["chapter"],
			image,
		)
	}

	stringifiedData, err := json.Marshal(images)
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

		manga := dbDriver.Product{
			Name:     name,
			Url:      url,
			Size:     0,
			SrcUrl:   "",
			AddDate:  time.Now(),
			UpdDate:  time.Now(),
			Titles:   []string{},
			Chapters: []dbDriver.Chapter{},
		}

		dbDriver.MangaCache.Add(manga)
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
		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		UpdateProductCache("manga", mux.Vars(r)["name"])

		apiGetMangaMain(w, r)
		mangaLoader.UpdateManga(mux.Vars(r)["name"])
	case "remove":
		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		dbDriver.MangaPagesCache.Remove(mux.Vars(r)["name"])
		UpdateProductPagesAllCache("manga", mux.Vars(r)["name"])

		dbDriver.RemoveManga(mux.Vars(r)["name"])
		apiGetMangaMain(w, r)
	case "changeName":
		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		UpdateProductCache("manga", mux.Vars(r)["name"])

		dbDriver.SetMangaName(mux.Vars(r)["name"], r.FormValue("name"))
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

		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		UpdateProductCache("manga", mux.Vars(r)["name"])

		dbDriver.AddMangaTitle(mux.Vars(r)["name"], header.Filename)
		apiGetMangaInfo(w, r)
	case "remTitle":
		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		UpdateProductCache("manga", mux.Vars(r)["name"])

		dbDriver.RemoveMangaTitle(mux.Vars(r)["name"], r.FormValue("fileName"))
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

		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		UpdateProductCache("manga", mux.Vars(r)["name"])

		dbDriver.SetMangaChapterName(mux.Vars(r)["name"], int(chapNumber), name)
	case "remove":
		chapNumber, err := strconv.ParseInt(mux.Vars(r)["chapter"], 10, 0)
		if err != nil {
			http.Error(w, "Incorrect request", 400)
		}

		dbDriver.MangaCache.Remove(mux.Vars(r)["name"])
		dbDriver.MangaPagesCache.RemoveChapter(mux.Vars(r)["name"], int(chapNumber))
		UpdateProductPagesCache("manga", mux.Vars(r)["name"], int(chapNumber))

		dbDriver.RemoveMangaChapter(mux.Vars(r)["name"], int(chapNumber))
	default:
		http.Error(w, "no such action", 400)
		return
	}

	apiGetMangaInfo(w, r)
}
