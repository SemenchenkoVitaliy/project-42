package mangaLoader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	dbDriver "project-42/mongoDriver"
)

func parseMangaType0(url string) dbDriver.Manga {
	// get manga url name and normal name
	mangaUrl := strings.Split(url, "/")[3]
	mangaName := strings.Title(strings.TrimSpace(strings.Replace(mangaUrl, "_", " ", -1)))

	// create new db manga object
	return dbDriver.Manga{
		Name:     mangaName,
		Url:      mangaUrl,
		Size:     0,
		SrcUrl:   url,
		AddDate:  time.Now(),
		UpdDate:  time.Now(),
		Titles:   []string{},
		Chapters: []dbDriver.MangaChapter{},
	}
}

func parseMangaType1(url string) dbDriver.Manga {
	// get manga url name and normal name
	mangaUrl := strings.Replace(url[strings.Index(url, "-")+1:strings.LastIndex(url, ".")], "-", "_", -1)
	mangaName := strings.Title(strings.TrimSpace(strings.Replace(mangaUrl, "_", " ", -1)))

	// create new db manga object
	return dbDriver.Manga{
		Name:     mangaName,
		Url:      mangaUrl,
		Size:     0,
		SrcUrl:   url,
		AddDate:  time.Now(),
		UpdDate:  time.Now(),
		Titles:   []string{},
		Chapters: []dbDriver.MangaChapter{},
	}
}

func parseChaptersType0(url string) []MangaChapter {
	// get html page
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	body := string(bytes)

	// get hostname
	hostname := url[(strings.Index(url, "//"))+2:]
	hostname = url[0:strings.Index(url, "//")+2] + hostname[0:strings.Index(hostname, "/")]

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

	result := []MangaChapter{}

	// push objects with chapters names and urls to result slice in reverse order
	for index, _ := range sliceBody {
		item := sliceBody[len(sliceBody)-index-1]
		if item == "" {
			continue
		}
		temp := strings.Split(item, "\" >")
		result = append(result, MangaChapter{Url: hostname + temp[0], Name: temp[1]})
	}

	return result
}

func parseChaptersType1(url string) []MangaChapter {
	// get html page
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	body := string(bytes)

	// get hostname
	hostname := url[(strings.Index(url, "//"))+2:]
	hostname = url[0:strings.Index(url, "//")+2] + hostname[0:strings.Index(hostname, "/")]

	// get slice of chapters urls
	start := strings.Index(body, "</style>") + 8
	body = body[start:]
	end := strings.Index(body, "<div style")
	body = body[0:end]
	body = strings.Replace(body, "&nbsp;&nbsp;", "", -1)
	sliceBody := strings.Split(body, "a href='")

	result := []MangaChapter{}

	// push objects with chapters names and urls to result slice in reverse order
	for index, _ := range sliceBody {
		item := sliceBody[len(sliceBody)-index-1]
		end := strings.Index(item, "</span>")

		if end == -1 {
			continue
		}

		temp := strings.Split(item[0:end], "' title=''>")
		result = append(result, MangaChapter{Url: hostname + temp[0], Name: temp[1]})
	}

	return result
}

func parseChapterType0(url string) []string {
	// get html page
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	body := string(bytes)

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

func parseChapterType1(url string) []string {
	// get html page
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	body := string(bytes)

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
