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

// getHostName accepts url and returns its hostname
func getHostname(url string) (hostname string) {
	hostname = url[(strings.Index(url, "//"))+2:]
	return url[0:strings.Index(url, "//")+2] + hostname[0:strings.Index(hostname, "/")]
}

// parseManga parses url and returns structure which can be added to database
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

// parseChapters loads and parses html page
//
// It returns slice of mangaChapter structs
func parseChapters(url string) (chapters []mangaChapter) {
	body, _ := load(url)
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

// parseChapters loads and parses html page
//
// It returns slice of images urls
func parseChapter(url string) (pages []string) {
	body, _ := load(url)

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
