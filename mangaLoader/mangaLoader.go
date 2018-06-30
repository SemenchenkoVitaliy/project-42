package mangaLoader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/dbDriver/mongo"
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

// stores master server and main database data structure objects
var (
	mainServer *netutils.Server
	db         *dbDriver.Database
)

// Init saves master server to which will be sent data to store and database
func Init(server *netutils.Server, database *dbDriver.Database) {
	mainServer = server
	db = database
}

// load makes http GET request
//
// It accepts url and returns data body, loaded by this url
func load(url string) (body string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	resp.Body.Close()
	return string(bytes), err
}

// loadChapter loads page html, parses it for images urls and loads those images
//
// It accepts url of page and directory path to load images into, returns slice
// of images filenames and error if any
func loadChapter(url, dir string) (imageNames []string, err error) {
	var name string
	body, _ := load(url)
	imagesUrls := parseChapter(url, body)
	imageNames = make([]string, 0, len(imagesUrls))

	for _, item := range imagesUrls {
		name = item[strings.LastIndex(item, "/")+1:]

		resp, err := http.Get(item)
		defer resp.Body.Close()
		if err != nil {
			utils.Log(err, fmt.Sprintf("get http page: %v", item))
			continue
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			utils.Log(err, fmt.Sprintf("convert http page to byte slice: %v", item))
			continue
		}

		mainServer.WriteFile(dir+name, bytes)
		imageNames = append(imageNames, name)
	}

	return imageNames, err
}

// UpdateManga loads the newest version of manga which provides site defined in manga SrcUrl
//
// It accepts manga Url and returns error if any
func UpdateManga(mangaName string) (err error) {
	manga, err := db.GetMangaSingle(mangaName)
	if err != nil {
		return err
	} else if manga.SrcUrl == "" {
		return fmt.Errorf("No source url")
	}

	body, _ := load(manga.SrcUrl)
	chapters := parseChapters(manga.SrcUrl, body)[manga.Size:]
	for index, item := range chapters {
		images, err := loadChapter(item.Url, fmt.Sprintf("/images/manga/%v/%v/", mangaName, index))
		if err != nil {
			return err
		}
		number := manga.Size + index
		chapter := dbDriver.Chapter{
			Name:   item.Name,
			Number: number,
		}
		if err = db.AddMangaChapter(mangaName, chapter); err != nil {
			return err
		}
		if err = db.AddMangaChapterPages(mangaName, number, images); err != nil {
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
	if err = db.AddManga(manga); err != nil {
		return err
	}
	return err
}
