package mangaLoader

import (
	"fmt"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/fsApi"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

func loadChapter(url, dir string) ([]string, error) {
	imageNames := []string{}
	images := []fsApi.UrlFile{}

	imageURLs := parseChapter(url)

	for _, item := range imageURLs {
		imageNames = append(imageNames, item[strings.LastIndex(item, "/")+1:])
		images = append(images, fsApi.UrlFile{Path: dir, Url: item})
	}

	if err := fsApi.MkDir(dir); err != nil {
		return imageNames, err
	}
	err := fsApi.LoadFiles(images)
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

func AddManga(url string) error {
	manga := parseManga(url)
	if err := dbDriver.AddManga(manga); err != nil {
		return err
	}
	if err := fsApi.MkDir("images/manga/" + manga.Url); err != nil {
		return err
	}
	return fsApi.MkDir("images/mangaTitles/" + manga.Url)
}
