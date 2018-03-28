package mangaLoader

import (
	"strings"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

type MangaChapter struct {
	Name string
	Url  string
}

// Load by url and parse html, get manga struct object
// Supported sites:
//   mangachan.me
//   readmanga.me
//   mintmanga.com
func parseManga(url string) dbDriver.Manga {
	// select transfer control to appropriate function depending on site
	if strings.Index(url, "readmanga.me") >= 0 {
		return parseMangaType0(url)
	} else if strings.Index(url, "mintmanga.com") >= 0 {
		return parseMangaType0(url)
	} else if strings.Index(url, "mangachan.me") >= 0 {
		return parseMangaType1(url)
	} else {
		return dbDriver.Manga{}
	}
}

// Load by url and parse html, get array of chapter struct objects
// Supported sites:
//   mangachan.me
//   readmanga.me
//   mintmanga.com
func parseChapters(url string) []MangaChapter {
	// select transfer control to appropriate function depending on site
	if strings.Index(url, "readmanga.me") >= 0 {
		return parseChaptersType0(url)
	} else if strings.Index(url, "mintmanga.com") >= 0 {
		return parseChaptersType0(url)
	} else if strings.Index(url, "mangachan.me") >= 0 {
		return parseChaptersType1(url)
	} else {
		return nil
	}
}

// Load by url and parse html, get array of chapter's urls of images
// Supported sites:
//   mangachan.me
//   readmanga.me
//   mintmanga.com
func parseChapter(url string) []string {
	// select transfer control to appropriate function depending on site
	if strings.Index(url, "readmanga.me") >= 0 {
		return parseChapterType0(url)
	} else if strings.Index(url, "mintmanga.com") >= 0 {
		return parseChapterType0(url)
	} else if strings.Index(url, "mangachan.me") >= 0 {
		return parseChapterType1(url)
	} else {
		return nil
	}
}
