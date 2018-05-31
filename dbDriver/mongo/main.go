package mongo

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/SemenchenkoVitaliy/project-42/cache"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

const (
	cacheSize int = 1000
	cacheTTL  int = 60
)

type Chapter struct {
	Name   string
	Number int
	Pages  []string
}

type Product struct {
	Size     int
	Url      string
	Name     string
	SrcUrl   string
	Covers   []string
	AddDate  time.Time
	UpdDate  time.Time
	Chapters []Chapter
}

type Database struct {
	manga       *mgo.Collection
	ranobe      *mgo.Collection
	mangaCache  *cache.Cache
	ranobeCache *cache.Cache
}

func NewDatabase() (db *Database) {
	return &Database{}
}

func (db *Database) Connect(user, password, dbName, ip string, port int) {
	session, err := mgo.Dial(fmt.Sprintf("%v:%v@%v:%v/%v",
		user,
		password,
		ip,
		port,
		dbName))
	if err != nil {
		panic(err)
		utils.LogCritical(err, "start MongoDB session")
	}
	session.SetMode(mgo.Monotonic, true)
	db.manga = session.DB(dbName).C("manga")
	db.ranobe = session.DB(dbName).C("ranobe")
	db.mangaCache = cache.NewCache(cacheSize, cacheTTL)
	db.ranobeCache = cache.NewCache(cacheSize, cacheTTL)
}

/******************************************************************************
	General
******************************************************************************/

func (db *Database) getProductUrls(c *mgo.Collection, start, quantity int, sort string) (urls []string, err error) {
	urlStructs := []struct {
		Url string
	}{}

	err = c.
		Find(bson.M{}).
		Sort(sort).
		Select(bson.M{
			"url": 1,
		}).
		Limit(quantity).
		Skip(start).
		All(&urlStructs)
	if err != nil {
		return urls, err
	}

	for _, item := range urlStructs {
		urls = append(urls, item.Url)
	}
	return urls, err
}

func (db *Database) getProductSingle(c *mgo.Collection, url string) (product Product, err error) {
	err = c.
		Find(bson.M{
			"url": url,
		}).
		One(&product)
	return product, err
}

func (db *Database) getProductMultiple(c *mgo.Collection, urls []string) (product []Product, err error) {
	err = c.
		Find(bson.M{
			"url": bson.M{
				"$in": urls,
			},
		}).
		All(&product)
	return product, err
}

func (db *Database) addProduct(c *mgo.Collection, product Product) (err error) {
	num, err := c.
		Find(bson.M{
			"url": product.Url,
		}).
		Count()
	if err != nil {
		return err
	}
	if num != 0 {
		return fmt.Errorf("product is already added")
	}
	return c.Insert(&product)
}

func (db *Database) removeProduct(c *mgo.Collection, url string) (err error) {
	return c.
		Remove(bson.M{
			"url": url,
		})
}

func (db *Database) addProductChapter(c *mgo.Collection, url string, chapter Chapter) (err error) {
	findSelector := bson.M{
		"url": url,
	}
	num, err := c.
		Find(findSelector).
		Count()
	if err != nil {
		return err
	}
	if num != 1 {
		return fmt.Errorf("manga is already added")
	}

	err = c.
		Update(findSelector, bson.M{
			"$push": bson.M{
				"chapters": chapter,
			},
		})
	if err != nil {
		return err
	}

	err = c.
		Update(findSelector, bson.M{
			"$set": bson.M{
				"upddate": time.Now(),
			},
		})
	if err != nil {
		return err
	}

	return c.
		Update(findSelector, bson.M{
			"$inc": bson.M{
				"size": 1,
			},
		})
}

func (db *Database) addProductChapterEmpty(c *mgo.Collection, url string, name string) (num int, err error) {
	var product Product
	err = c.
		Find(bson.M{
			"url": url,
		}).
		One(&product)
	if err != nil {
		return num, err
	}
	num = product.Size
	return num, db.addProductChapter(c, url, Chapter{
		Name:   name,
		Number: num,
		Pages:  []string{},
	})
}

func (db *Database) removeProductChapter(c *mgo.Collection, url string, number int) (err error) {
	findSelector := bson.M{
		"url": url,
	}

	err = c.
		Update(findSelector, bson.M{
			"$pull": bson.M{
				"chapters": bson.M{
					"number": number,
				},
			},
		})
	if err != nil {
		return err
	}

	err = c.
		Update(findSelector, bson.M{
			"$set": bson.M{
				"upddate": time.Now(),
			},
		})
	if err != nil {
		return err
	}

	return c.
		Update(findSelector, bson.M{
			"$inc": bson.M{
				"size": -1,
			},
		})
}

func (db *Database) addProductChapterPages(c *mgo.Collection, url string, number int, pages []string) (err error) {
	return c.
		Update(bson.M{
			"url":             url,
			"chapters.number": number,
		}, bson.M{
			"$push": bson.M{
				"chapters.$.pages": bson.M{
					"$each": pages,
				},
			},
		})
}

func (db *Database) removeProductChapterPages(c *mgo.Collection, url string, number int) (err error) {
	return c.
		Update(bson.M{
			"url":             url,
			"chapters.number": number,
		}, bson.M{
			"$set": bson.M{
				"chapters.$.pages": []string{},
			},
		})
}

func (db *Database) addProductCover(c *mgo.Collection, url, coverImageName string) (err error) {
	db.removeProductCover(c, url, coverImageName)
	return c.
		Update(bson.M{
			"url": url,
		}, bson.M{
			"$push": bson.M{
				"covers": coverImageName,
			},
		})
}

func (db *Database) removeProductCover(c *mgo.Collection, url, coverImageName string) (err error) {
	return c.
		Update(bson.M{
			"url": url,
		}, bson.M{
			"$pull": bson.M{
				"covers": coverImageName,
			},
		})
}

func (db *Database) setProductName(c *mgo.Collection, url, newName string) (err error) {
	return c.
		Update(bson.M{
			"url": url,
		}, bson.M{
			"$set": bson.M{
				"name": newName,
			},
		})
}

func (db *Database) setProductChapterName(c *mgo.Collection, url string, number int, newName string) (err error) {
	return c.
		Update(bson.M{
			"url":             url,
			"chapters.number": number,
		}, bson.M{
			"$set": bson.M{
				"chapters.$.name": newName,
			},
		})
}

/******************************************************************************
	Manga
******************************************************************************/

func (db *Database) GetMangaUrls(start, quantity int, sort string) (urls []string, err error) {
	return db.getProductUrls(db.manga, start, quantity, sort)
}

func (db *Database) GetMangaSingle(url string) (manga Product, err error) {
	if item, ok := db.mangaCache.Get(url); ok {
		return item.(Product), nil
	} else {
		manga, err = db.getProductSingle(db.manga, url)
		if err == nil {
			db.mangaCache.Add(url, manga)
		}
		return manga, err
	}
}

func (db *Database) GetMangaMultiple(urls []string) (mangaArr []Product, err error) {
	for _, url := range urls {
		if item, ok := db.mangaCache.Get(url); ok {
			mangaArr = append(mangaArr, item.(Product))
		} else if manga, err := db.getProductSingle(db.manga, url); err == nil {
			db.mangaCache.Add(url, manga)
			mangaArr = append(mangaArr, manga)
		}
	}
	return mangaArr, err
}

func (db *Database) AddManga(manga Product) (err error) {
	db.mangaCache.Add(manga.Url, manga)
	return db.addProduct(db.manga, manga)
}

func (db *Database) RemoveManga(url string) (err error) {
	db.mangaCache.Delete(url)
	return db.removeProduct(db.manga, url)
}

func (db *Database) AddMangaChapter(url string, chapter Chapter) (err error) {
	db.mangaCache.Delete(url)
	return db.addProductChapter(db.manga, url, chapter)
}

func (db *Database) AddMangaChapterEmpty(url, name string) (num int, err error) {
	db.mangaCache.Delete(url)
	return db.addProductChapterEmpty(db.manga, url, name)
}

func (db *Database) RemoveMangaChapter(url string, number int) (err error) {
	db.mangaCache.Delete(url)
	return db.removeProductChapter(db.manga, url, number)
}

func (db *Database) AddMangaChapterPages(url string, number int, pages []string) (err error) {
	db.mangaCache.Delete(url)
	return db.addProductChapterPages(db.manga, url, number, pages)
}

func (db *Database) RemoveMangaChapterPages(url string, number int) (err error) {
	db.mangaCache.Delete(url)
	return db.removeProductChapterPages(db.manga, url, number)
}

func (db *Database) AddMangaCover(url, coverImageName string) (err error) {
	db.mangaCache.Delete(url)
	return db.addProductCover(db.manga, url, coverImageName)
}

func (db *Database) RemoveMangaCover(url, coverImageName string) (err error) {
	db.mangaCache.Delete(url)
	return db.removeProductCover(db.manga, url, coverImageName)
}

func (db *Database) SetMangaName(url, newName string) (err error) {
	db.mangaCache.Delete(url)
	return db.setProductName(db.manga, url, newName)
}

func (db *Database) SetMangaChapterName(url string, number int, newName string) (err error) {
	db.mangaCache.Delete(url)
	return db.setProductChapterName(db.manga, url, number, newName)
}

func (db *Database) RemoveFromMangaCache(url string) {
	db.mangaCache.Delete(url)
}

/******************************************************************************
	Ranobe
******************************************************************************/

func (db *Database) GetRanobeUrls(start, quantity int, sort string) (urls []string, err error) {
	return db.getProductUrls(db.ranobe, start, quantity, sort)
}

func (db *Database) GetRanobeSingle(url string) (ranobe Product, err error) {
	if item, ok := db.ranobeCache.Get(url); ok {
		return item.(Product), nil
	} else {
		ranobe, err = db.getProductSingle(db.ranobe, url)
		if err == nil {
			db.ranobeCache.Add(url, ranobe)
		}
		return ranobe, err
	}
}

func (db *Database) GetRanobeMultiple(urls []string) (ranobeArr []Product, err error) {
	ranobeArr = make([]Product, len(urls))
	for _, url := range urls {
		if item, ok := db.ranobeCache.Get(url); ok {
			ranobeArr = append(ranobeArr, item.(Product))
		} else if ranobe, err := db.getProductSingle(db.ranobe, url); err == nil {
			db.ranobeCache.Add(url, ranobe)
			ranobeArr = append(ranobeArr, ranobe)
		}
	}
	return ranobeArr, err
}

func (db *Database) AddRanobe(ranobe Product) (err error) {
	db.ranobeCache.Add(ranobe.Url, ranobe)
	return db.addProduct(db.ranobe, ranobe)
}

func (db *Database) RemoveRanobe(url string) (err error) {
	db.ranobeCache.Delete(url)
	return db.removeProduct(db.ranobe, url)
}

func (db *Database) AddRanobeChapter(url string, chapter Chapter) (err error) {
	db.ranobeCache.Delete(url)
	return db.addProductChapter(db.ranobe, url, chapter)
}

func (db *Database) AddRanobeChapterEmpty(url, name string) (num int, err error) {
	db.ranobeCache.Delete(url)
	return db.addProductChapterEmpty(db.ranobe, url, name)
}

func (db *Database) RemoveRanobeChapter(url string, number int) (err error) {
	db.ranobeCache.Delete(url)
	return db.removeProductChapter(db.ranobe, url, number)
}

func (db *Database) AddRanobeChapterPages(url string, number int, pages []string) (err error) {
	db.ranobeCache.Delete(url)
	return db.addProductChapterPages(db.ranobe, url, number, pages)
}

func (db *Database) RemoveRanobeChapterPages(url string, number int) (err error) {
	db.ranobeCache.Delete(url)
	return db.removeProductChapterPages(db.ranobe, url, number)
}

func (db *Database) AddRanobeCover(url, coverImageName string) (err error) {
	db.ranobeCache.Delete(url)
	return db.addProductCover(db.ranobe, url, coverImageName)
}

func (db *Database) RemoveRanobeCover(url, coverImageName string) (err error) {
	db.ranobeCache.Delete(url)
	return db.removeProductCover(db.ranobe, url, coverImageName)
}

func (db *Database) SetRanobeName(url, newName string) (err error) {
	db.ranobeCache.Delete(url)
	return db.setProductName(db.ranobe, url, newName)
}

func (db *Database) SetRanobeChapterName(url string, number int, newName string) (err error) {
	db.ranobeCache.Delete(url)
	return db.setProductChapterName(db.ranobe, url, number, newName)
}

func (db *Database) RemoveFromRanobeCache(url string) {
	db.ranobeCache.Delete(url)
}
