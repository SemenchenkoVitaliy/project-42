package mangaLoader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

var mainServer tcp.Server

func Init(server tcp.Server) {
	mainServer = server
}

type writeFileData struct {
	Path string
	Data []byte
}

func WriteFile(path string, fileData []byte) {
	data := writeFileData{
		Path: path,
		Data: fileData,
	}
	b, _ := json.Marshal(data)
	mainServer.Send(b, 2)
}

func MkDir(path string) {
	mainServer.Send([]byte(path), 3)
}

func loadChapter(url, dir string) (imageNames []string, err error) {
	var name string
	imagesUrls := parseChapter(url)

	for _, item := range imagesUrls {
		name = item[strings.LastIndex(item, "/")+1:]

		resp, err := http.Get(item)
		defer resp.Body.Close()
		if err != nil {
			common.CreateLog(err, fmt.Sprintf("get http page: %v", item))
			continue
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			common.CreateLog(err, fmt.Sprintf("convert http page to byte slice: %v", item))
			continue
		}

		WriteFile("/"+dir+"/"+name, bytes)
		imageNames = append(imageNames, name)
	}

	return imageNames, err
}

func UpdateManga(mangaName string) error {
	manga, err := dbDriver.GetManga(mangaName)
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

		chapter := dbDriver.MangaChapter{
			Name:   item.Name,
			Number: manga.Size + index,
		}
		if err = dbDriver.AddMangaChapter(mangaName, chapter, images); err != nil {
			return err
		}
	}
	return nil
}

func AddManga(url string) (err error) {
	manga := parseManga(url)
	if err = dbDriver.AddManga(manga); err != nil {
		return err
	}
	MkDir("images/manga/" + manga.Url)
	MkDir("images/mangaTitles/" + manga.Url)
	return err
}
