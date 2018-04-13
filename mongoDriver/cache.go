package mongoDriver

import (
	"sort"
	"sync"
)

var (
	MangaCache      = NewProductCache()
	MangaPagesCache = NewProductPageCache()
)

type ProductPageCache struct {
	cache map[string]map[int][]string
	sync.Mutex
}

func NewProductPageCache() *ProductPageCache {
	return &ProductPageCache{cache: make(map[string]map[int][]string)}
}

func (c *ProductPageCache) Add(name string, chapter int, pages []string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.cache[name]; !ok {
		c.cache[name] = make(map[int][]string)
	}
	sort.Strings(pages)
	c.cache[name][chapter] = append([]string{}, pages...)
}

func (c *ProductPageCache) Find(name string, chapter int) (pages []string, ok bool) {
	c.Lock()
	defer c.Unlock()

	if _, ok = c.cache[name]; ok {
		if p, ok := c.cache[name][chapter]; ok {
			pages = append(pages, p...)
			return pages, true
		}
	}
	return pages, false
}

func (c *ProductPageCache) Remove(name string) {
	c.Lock()
	defer c.Unlock()

	delete(c.cache, name)
}

func (c *ProductPageCache) RemoveChapter(name string, chapter int) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.cache[name]; ok {
		delete(c.cache[name], chapter)
	}
}

type ProductCache struct {
	cache map[string]Product
	sync.Mutex
}

func NewProductCache() *ProductCache {
	return &ProductCache{cache: make(map[string]Product)}
}

func (c *ProductCache) Add(product Product) {
	c.Lock()
	defer c.Unlock()

	titles := []string{}
	chapters := []Chapter{}
	titles = append(titles, product.Titles...)
	chapters = append(chapters, product.Chapters...)

	c.cache[product.Url] = Product{
		Size:     product.Size,
		Url:      product.Url,
		Name:     product.Name,
		SrcUrl:   product.SrcUrl,
		Titles:   titles,
		AddDate:  product.AddDate,
		UpdDate:  product.UpdDate,
		Chapters: chapters,
	}
}

func (c ProductCache) Find(name string) (product Product, ok bool) {
	c.Lock()
	defer c.Unlock()

	if _, ok = c.cache[name]; ok {
		product = Product{
			Size:     c.cache[name].Size,
			Url:      c.cache[name].Url,
			Name:     c.cache[name].Name,
			SrcUrl:   c.cache[name].SrcUrl,
			Titles:   []string{},
			AddDate:  c.cache[name].AddDate,
			UpdDate:  c.cache[name].UpdDate,
			Chapters: []Chapter{},
		}
		product.Titles = append(product.Titles, c.cache[name].Titles...)
		product.Chapters = append(product.Chapters, c.cache[name].Chapters...)
		return product, true
	}
	return product, false
}

func (c *ProductCache) Remove(name string) {
	c.Lock()
	defer c.Unlock()

	delete(c.cache, name)
}
