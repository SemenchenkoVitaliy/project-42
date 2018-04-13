package mongoDriver

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

var (
	session               *mgo.Session
	mangaCollection       *mgo.Collection
	mangaImagesCollection *mgo.Collection
)

func init() {
	url := fmt.Sprintf("%v:%v", common.Config.Db.HostIP, common.Config.Db.HostPort)
	session, err := mgo.Dial(url)
	if err != nil {
		common.CreateLogCritical(err, "start MongoDB session")
	}

	if err = session.DB(common.Config.Db.DbName).Login(common.Config.Db.User, common.Config.Db.Password); err != nil {
		common.CreateLogCritical(err, "authenticate MongoDB session")
	}

	session.SetMode(mgo.Monotonic, true)

	mangaCollection = session.DB(common.Config.Db.DbName).C("mangalist")
	mangaImagesCollection = session.DB(common.Config.Db.DbName).C("mangaimages")
}

type Chapter struct {
	Name   string
	Number int
}

type Product struct {
	Size     int
	Url      string
	Name     string
	SrcUrl   string
	Titles   []string
	AddDate  time.Time
	UpdDate  time.Time
	Chapters []Chapter
}

func GetMangaUrls(start, quantity int, sort string) (urls []string, err error) {
	urlStructs := []struct {
		Url string
	}{}

	err = mangaCollection.
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

func GetMangaSingle(url string) (manga Product, err error) {
	err = mangaCollection.
		Find(bson.M{
			"url": url,
		}).
		One(&manga)
	return manga, err
}

func GetMangaMultiple(urls []string) (manga []Product, err error) {
	err = mangaCollection.
		Find(bson.M{
			"url": bson.M{
				"$in": urls,
			},
		}).
		All(&manga)
	return manga, err
}

func GetRanobeUrls(start, quantity int, sort string) (urls []string, err error) {
	return urls, err
}

func GetRanobeSingle(url string) (manga Product, err error) {
	return manga, err
}

func GetRanobeMultiple(urls []string) (manga []Product, err error) {
	return manga, err
}

func GetMangaChapterPages(url string, chapter int) (pages []string, err error) {
	images := []struct {
		Image string
	}{}

	err = mangaImagesCollection.
		Find(bson.M{
			"manga":  url,
			"number": chapter,
		}).
		Sort("image").
		Select(bson.M{
			"image": 1,
		}).
		All(&images)
	if err != nil {
		return pages, err
	}
	for _, item := range images {
		pages = append(pages, item.Image)
	}
	return pages, err
}

func AddManga(manga Product) (err error) {
	num, err := mangaCollection.
		Find(bson.M{
			"url": manga.Url,
		}).
		Count()
	if err != nil {
		return err
	}
	if num != 0 {
		return fmt.Errorf("manga is already added")
	}
	return mangaCollection.Insert(&manga)
}

func RemoveManga(url string) (err error) {
	err = mangaCollection.
		Remove(bson.M{
			"url": url,
		})
	if err != nil {
		return err
	}

	_, err = mangaImagesCollection.
		RemoveAll(bson.M{
			"manga": url,
		})
	return err
}

func AddMangaChapter(url string, chapter Chapter) (err error) {
	findSelector := bson.M{
		"url": url,
	}
	num, err := mangaCollection.
		Find(findSelector).
		Count()
	if err != nil {
		return err
	}
	if num != 1 {
		return fmt.Errorf("manga is already added")
	}

	err = mangaCollection.
		Update(findSelector, bson.M{
			"$push": bson.M{
				"chapters": chapter,
			},
		})
	if err != nil {
		return err
	}

	err = mangaCollection.
		Update(findSelector, bson.M{
			"$set": bson.M{
				"upddate": time.Now(),
			},
		})
	if err != nil {
		return err
	}

	return mangaCollection.
		Update(findSelector, bson.M{
			"$inc": bson.M{
				"size": 1,
			},
		})
}

func RemoveMangaChapter(url string, number int) (err error) {
	findSelector := bson.M{
		"url": url,
	}

	err = mangaCollection.
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

	err = mangaCollection.
		Update(findSelector, bson.M{
			"$set": bson.M{
				"upddate": time.Now(),
			},
		})
	if err != nil {
		return err
	}

	return mangaCollection.
		Update(findSelector, bson.M{
			"$inc": bson.M{
				"size": -1,
			},
		})
}

func AddMangaChapterPages(name string, number int, pages []string) (err error) {
	type imageStruct struct {
		Manga  string
		Number int
		Image  string
	}

	var images []interface{}

	for _, page := range pages {
		images = append(images, imageStruct{
			Manga:  name,
			Number: number,
			Image:  page,
		})
	}

	return mangaImagesCollection.
		Insert(images...)
}

func RemoveMangaChapterPages(name string, number int) (err error) {
	return mangaImagesCollection.
		Remove(bson.M{
			"manga":  name,
			"number": number,
		})
}

func AddMangaTitle(name, titleName string) (err error) {
	return mangaCollection.
		Update(bson.M{
			"url": name,
		}, bson.M{
			"$push": bson.M{
				"titles": titleName,
			},
		})
}

func RemoveMangaTitle(name, title string) (err error) {
	return mangaCollection.
		Update(bson.M{
			"url": name,
		}, bson.M{
			"$pull": bson.M{
				"titles": title,
			},
		})
}

func SetMangaName(name, newName string) (err error) {
	return mangaCollection.
		Update(bson.M{
			"url": name,
		}, bson.M{
			"$set": bson.M{
				"name": newName,
			},
		})
}

func SetMangaChapterName(name string, number int, newName string) (err error) {
	return mangaCollection.
		Update(bson.M{
			"url":             name,
			"chapters.number": number,
		}, bson.M{
			"$set": bson.M{
				"chapters.$.name": newName,
			},
		})
}
