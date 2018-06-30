package mangaLoader

import (
	"fmt"
	"strings"
	"time"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/dbDriver/mongo"
)

// mangaChapter represents name of manga and url of page where it can be
// downladed
type mangaChapter struct {
	Name string
	Url  string
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
		Covers:   []string{},
		Chapters: []dbDriver.Chapter{},
	}
}

// parseChapters loads and parses html page
//
// It returns slice of mangaChapter structs
func parseChapters(url, body string) (chapters []mangaChapter) {
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
func parseChapter(url, body string) (pages []string) {
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

// parseMangaType0 is realisation of parseManga for sites with layout
// similar to 'http://readmanga.me' and 'http://mintmanga.com'
func parseMangaType0(url string) (mangaUrl, mangaName string) {
	mangaUrl = strings.Split(url, "/")[3]
	mangaName = strings.Title(strings.TrimSpace(strings.Replace(mangaUrl, "_", " ", -1)))
	return mangaUrl, mangaName
}

// parseMangaType1 is realisation of parseManga for sites with layout
// similar to 'http://mangachan.me'
func parseMangaType1(url string) (mangaUrl, mangaName string) {
	mangaUrl = strings.Replace(url[strings.Index(url, "-")+1:strings.LastIndex(url, ".")], "-", "_", -1)
	mangaName = strings.Title(strings.TrimSpace(strings.Replace(mangaUrl, "_", " ", -1)))
	return mangaUrl, mangaName
}

// parseChaptersType0 is realisation of parseChapters for sites with layout
// similar to 'http://readmanga.me' and 'http://mintmanga.com'
func parseChaptersType0(hostname, body string) (chapters []mangaChapter) {
	// get slice of chapters urls
	start := strings.Index(body, "class=\"form-control\">") + 21
	body = body[start:]
	end := strings.Index(body, "</select>")
	body = body[0:end]
	body = strings.Replace(body, "  ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "selected=&quot;selected&quot;", "", -1)
	body = strings.Replace(body, "<option value=\"", "", -1)
	sliceBody := strings.Split(body, "</option>")

	// push objects with chapters names and urls to result slice in reverse order
	for index, _ := range sliceBody {
		item := sliceBody[len(sliceBody)-index-1]
		if item == "" {
			continue
		}
		temp := strings.Split(item, "\" >")
		chapters = append(chapters, mangaChapter{Url: hostname + temp[0], Name: temp[1]})
	}

	return chapters
}

// parseChaptersType1 is realisation of parseManga for sites with layout
// similar to 'http://mangachan.me'
func parseChaptersType1(hostname, body string) (chapters []mangaChapter) {
	// get slice of chapters urls
	start := strings.Index(body, "</style>") + 8
	body = body[start:]
	end := strings.Index(body, "<div style")
	body = body[0:end]
	body = strings.Replace(body, "&nbsp;&nbsp;", "", -1)
	sliceBody := strings.Split(body, "a href='")

	for index, _ := range sliceBody {
		item := sliceBody[len(sliceBody)-index-1]
		end := strings.Index(item, "</span>")

		if end == -1 {
			continue
		}

		temp := strings.Split(item[0:end], "' title=''>")
		chapters = append(chapters, mangaChapter{Url: hostname + temp[0], Name: temp[1]})
	}

	return chapters
}

// parseChapterType0 is realisation of parseChapter for sites with layout
// similar to 'http://readmanga.me' and 'http://mintmanga.com'
func parseChapterType0(url, body string) []string {
	// get slice of strings of parts of images absolute paths
	start := strings.Index(body, "rm_h.init(") + 10
	body = body[start:]
	end := strings.Index(body, "], 0, false);")
	body = body[0:end]
	body = strings.Replace(body, "'", "", -1)
	body = strings.Replace(body, "\"", "", -1)
	body = strings.Replace(body, "],[", ";", -1)
	body = strings.Replace(body, "[", "", -1)
	body = strings.Replace(body, "]", "", -1)
	sliceBody := strings.Split(body, ";")

	result := []string{}

	// push images urls to result slice
	for _, item := range sliceBody {
		temp := strings.Split(item, ",")
		if temp[1] != "" {
			result = append(result, temp[1]+temp[2])
		} else {
			fmt.Println("Censored: " + url)
		}
	}

	return result
}

// parseChapterType1 is realisation of parseChapter for sites with layout
// similar to 'http://mangachan.me'
func parseChapterType1(body string) []string {
	// get slice of strings of parts of images absolute paths
	start := strings.Index(body, "\"fullimg\":[") + 11
	body = body[start:]
	end := strings.Index(body, "]")
	body = body[0:end]
	body = strings.Replace(body, "\"", "", -1)
	sliceBody := strings.Split(body, ",")

	result := []string{}

	// push images urls to result slice
	for _, item := range sliceBody {
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
