package mongoDriver

import (
	"sort"
)

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
	c.Cache[manga.Url] = manga
}

func (c *mangCache) Find(name string) (Manga, bool) {
	if _, ok := c.Cache[name]; ok {
		return c.Cache[name], true
	}
	return Manga{}, false
}

func (c *mangCache) Remove(name string) {
	delete(c.Cache, name)
}
