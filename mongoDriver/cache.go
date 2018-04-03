package mongoDriver

import (
	"sort"
)

var (
	imageCache mangaImgCache
	mangaCache mangCache
)

func init() {
	imageCache.Cache = make(map[string]map[int][]string)
	mangaCache.Cache = make(map[string]Manga)
}

type mangaImgCache struct {
	Cache map[string]map[int][]string
}

func (c *mangaImgCache) Add(manga string, chapter int, images []string) {
	if _, ok := c.Cache[manga]; !ok {
		c.Cache[manga] = make(map[int][]string)
	}
	sort.Strings(images)
	c.Cache[manga][chapter] = images
}

func (c *mangaImgCache) Find(manga string, chapter int) ([]string, bool) {
	if _, ok := c.Cache[manga]; ok {
		if _, ok = c.Cache[manga][chapter]; ok {
			return c.Cache[manga][chapter], true
		}
	}
	return []string{}, false
}

func (c *mangaImgCache) Remove(manga string) {
	delete(c.Cache, manga)
}

type mangCache struct {
	Cache map[string]Manga
	count uint
}

func (c *mangCache) Add(manga Manga) {
	titles := []string{}
	chapters := []MangaChapter{}
	titles = append(titles, manga.Titles...)
	chapters = append(chapters, manga.Chapters...)

	c.Cache[manga.Url] = Manga{
		Size:     manga.Size,
		Url:      manga.Url,
		Name:     manga.Name,
		SrcUrl:   manga.SrcUrl,
		Titles:   titles,
		AddDate:  manga.AddDate,
		UpdDate:  manga.UpdDate,
		Chapters: chapters,
	}
}

func (c mangCache) Find(name string) (Manga, bool) {
	if _, ok := c.Cache[name]; ok {
		result := Manga{
			Size:     c.Cache[name].Size,
			Url:      c.Cache[name].Url,
			Name:     c.Cache[name].Name,
			SrcUrl:   c.Cache[name].SrcUrl,
			Titles:   []string{},
			AddDate:  c.Cache[name].AddDate,
			UpdDate:  c.Cache[name].UpdDate,
			Chapters: []MangaChapter{},
		}
		result.Titles = append(result.Titles, c.Cache[name].Titles...)
		result.Chapters = append(result.Chapters, c.Cache[name].Chapters...)
		return result, true
	}
	return Manga{}, false
}

func (c *mangCache) Remove(name string) {
	delete(c.Cache, name)
}
