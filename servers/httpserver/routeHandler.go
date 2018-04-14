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

func httpAdmin(w http.ResponseWriter, r *http.Request) {
	mangaUrls, err := dbDriver.GetMangaUrls(0, 10, "name")
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
		Manga     []dbDriver.Product
		Ranobe    []dbDriver.Product
		PublicUrl string
	}{
		Manga:     manga,
		Ranobe:    ranobe,
		PublicUrl: common.Config.PublicUrl,
	}

	err = templates.ExecuteTemplate(w, "admin", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template admin")
		return
	}
}

func httpAdminManga(w http.ResponseWriter, r *http.Request) {
	manga, err := dbDriver.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}

	data := struct {
		Manga     dbDriver.Product
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

	err = templates.ExecuteTemplate(w, "main", data)
	if err != nil {
		writeServerInternalError(w, err, "Execute template main")
		return
	}
}

func httpMangaMain(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	mangaUrls, err := dbDriver.GetMangaUrls(page*20, 20, "name")
	if err != nil {
		writeServerInternalError(w, err, "Get manga urls")
		return
	}

	data := make([]dbDriver.Product, 0, 20)

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
		data = append(data, product)
	}

	if len(data) == 0 {
		writeServerInternalError(w, err, "Get multiple manga")
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
	data, err := dbDriver.GetMangaSingle(mux.Vars(r)["name"])
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

	manga, err := dbDriver.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		writeServerInternalError(w, err, "Get manga "+mux.Vars(r)["name"])
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

	data := struct {
		Manga          dbDriver.Product
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
