package mangaLoader

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

type mangaChapter struct {
	Name string
	Url  string
}

func load(url string) (body string) {
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(bytes)
}

func getHostname(url string) (hostname string) {
	hostname = url[(strings.Index(url, "//"))+2:]
	return url[0:strings.Index(url, "//")+2] + hostname[0:strings.Index(hostname, "/")]
}

// parseManga parses url and creates
// structure to add to database
func parseManga(url string) (manga dbDriver.Product) {
	var mangaUrl string
	var mangaName string

	switch getHostname(url) {
	case "http://readmanga.me":
		mangaUrl, mangaName = parseMangaType0(url)
	case "http://mintmanga.com":
		mangaUrl, mangaName = parseMangaType0(url)
	case "http://mangachan.me":
		mangaUrl, mangaName = parseMangaType1(url)
	default:
		return manga
	}

	return dbDriver.Product{
		Name:     mangaName,
		Url:      mangaUrl,
		Size:     0,
		SrcUrl:   url,
		AddDate:  time.Now(),
		UpdDate:  time.Now(),
		Titles:   []string{},
		Chapters: []dbDriver.Chapter{},
	}
}

// parseChapters loads and parses html page and
// returns slice of structs with name and url
func parseChapters(url string) (chapters []mangaChapter) {
	body := load(url)
	hostname := getHostname(url)

	switch hostname {
	case "http://readmanga.me":
		return parseChaptersType0(hostname, body)
	case "http://mintmanga.com":
		return parseChaptersType0(hostname, body)
	case "http://mangachan.me":
		return parseChaptersType1(hostname, body)
	default:
		return chapters
	}
}

// parseChapters loads and parses html page and
// returns slice of pages urls
func parseChapter(url string) (pages []string) {
	body := load(url)

	switch getHostname(url) {
	case "http://readmanga.me":
		return parseChapterType0(body, url)
	case "http://mintmanga.com":
		return parseChapterType0(body, url)
	case "http://mangachan.me":
		return parseChapterType1(body)
	default:
		return pages
	}
}
