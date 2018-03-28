package mangaLoader

import (
	"fmt"
	"strconv"

	"github.com/SemenchenkoVitaliy/project-42/fsApi"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

func LoadChapter(url, dir string) []string {
	images := parseChapter(url)
	var names []string
	var name string

	arr := []fsApi.UrlFile{}

	for _, item := range images {
		arr = append(arr, fsApi.UrlFile{Path: dir + "/" + name, Url: item})
	}
	fsApi.LoadFiles(arr)

	return names
}

func LoadManga(mangaName string) error {
	manga, err := dbDriver.GetManga(mangaName)
	if err != nil {
		return err
	}
	if manga.SrcUrl == "" {
		return fmt.Errorf("No source url")
	}
	chapters := parseChapters(manga.SrcUrl)[manga.Size:]
	var curDir string

	for index, item := range chapters {
		curDir = "images/manga/" + mangaName + "/" + strconv.Itoa(index)

		fsApi.MkDir(curDir)
		images := LoadChapter(item.Url, curDir)

		dbDriver.AddMangaChapter(mangaName, dbDriver.MangaChapter{Name: item.Name, Number: manga.Size + index}, images)
	}
	return nil
}

func AddManga(url string) error {
	manga := parseManga(url)

	err := dbDriver.AddManga(manga)
	if err != nil {
		return err
	}

	fsApi.MkDir("images/manga/" + manga.Url)
	fsApi.MkDir("images/mangaTitles/" + manga.Url)

	return nil
}
