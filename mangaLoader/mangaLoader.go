package mangaLoader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

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
func load(url string) (body []byte, err error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return body, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return bytes, err
}

// loadChapter loads page html, parses it for images urls and loads those images
//
// It accepts url of page and directory path to load images into, returns slice
// of images filenames and error if any
func loadChapter(url, dir string) (imageNames []string, err error) {
	body, err := load(url)
	if err != nil {
		panic(err)
	}
	imagesUrls := parseChapter(url, string(body))
	imageNames = make([]string, len(imagesUrls))

	var wg sync.WaitGroup
	wg.Add(len(imagesUrls))

	for index, imageUrl := range imagesUrls {
		go func(index int, url string) {
			data, err := load(url)
			if err != nil {
				utils.Log(err, fmt.Sprintf("load http page: %v", url))
				return
			}

			imageNames[index] = url[strings.LastIndex(url, "/")+1:]
			mainServer.WriteFile(dir+imageNames[index], data)
			wg.Done()
		}(index, imageUrl)
	}

	wg.Wait()
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

	body, err := load(manga.SrcUrl)
	if err != nil {
		return err
	}
	chapters := parseChapters(manga.SrcUrl, string(body))[manga.Size:]
	for index, item := range chapters {
		images, err := loadChapter(item.Url, fmt.Sprintf("/images/manga/%v/%v/", mangaName, manga.Size+index))
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
