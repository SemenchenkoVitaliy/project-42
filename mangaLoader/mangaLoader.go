// Package mangaLoader contains tools for loading and updating manga from other
// sources
package mangaLoader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

// stores master server data structure
var mainServer *tcp.Server

// Init saves master server to which will be sent writeFile and mkDir requests in a local variable
func Init(server *tcp.Server) {
	mainServer = server
}

// loadChapter loads page html, parses it for images urls and loads those images
//
// It accepts url of page and directory path to load images into, returns slice
// of images filenames and error if any
func loadChapter(url, dir string) (imageNames []string, err error) {
	var name string
	imagesUrls := parseChapter(url)
	imageNames = make([]string, 0, len(imagesUrls))

	for _, item := range imagesUrls {
		name = item[strings.LastIndex(item, "/")+1:]

		resp, err := http.Get(item)
		defer resp.Body.Close()
		if err != nil {
			common.Log(err, fmt.Sprintf("get http page: %v", item))
			continue
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			common.Log(err, fmt.Sprintf("convert http page to byte slice: %v", item))
			continue
		}

		mainServer.WriteFile("/"+dir+"/"+name, bytes)
		imageNames = append(imageNames, name)
	}

	return imageNames, err
}

// UpdateManga loads the newest version of manga which provides site defined in manga SrcUrl
//
// It accepts manga Url and returns error if any
func UpdateManga(mangaName string) (err error) {
	manga, err := dbDriver.GetMangaSingle(mangaName)
	if err != nil {
		return err
	} else if manga.SrcUrl == "" {
		return fmt.Errorf("No source url")
	}

	chapters := parseChapters(manga.SrcUrl)[manga.Size:]
	for index, item := range chapters {
		images, err := loadChapter(item.Url, fmt.Sprintf("images/manga/%v/%v/", mangaName, index))
		if err != nil {
			return err
		}
		number := manga.Size + index
		chapter := dbDriver.Chapter{
			Name:   item.Name,
			Number: number,
		}
		if err = dbDriver.AddMangaChapter(mangaName, chapter); err != nil {
			return err
		}
		if err = dbDriver.AddMangaChapterPages(mangaName, number, images); err != nil {
			return err
		}
	}
	return err
}

// AddManga adds manga to database and creates directories for it
//
// It accepts url of site which will be used in future to load new verions and
// returns error if any
func AddManga(url string) (err error) {
	manga := parseManga(url)
	if err = dbDriver.AddManga(manga); err != nil {
		return err
	}
	mainServer.MkDir("images/manga/" + manga.Url)
	mainServer.MkDir("images/mangaTitles/" + manga.Url)
	return err
}
