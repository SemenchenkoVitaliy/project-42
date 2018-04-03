package mongoDriver

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

type MangaImage struct {
	Manga  string
	Number int
	Image  string
}

type MangaChapter struct {
	Name   string
	Number int
}

type MangaMin struct {
	Size    int
	Url     string
	Name    string
	AddDate time.Time
	UpdDate time.Time
}

type Manga struct {
	Size     int
	Url      string
	Name     string
	SrcUrl   string
	Titles   []string
	AddDate  time.Time
	UpdDate  time.Time
	Chapters []MangaChapter
}

type RanobeChapter struct {
	Name   string
	Number int
}

type Ranobe struct {
	Size     int
	Url      string
	Name     string
	SrcUrl   string
	Titles   []string
	AddDate  time.Time
	UpdDate  time.Time
	Chapters []RanobeChapter
}

var (
	session               *mgo.Session
	mangaCollection       *mgo.Collection
	mangaImagesCollection *mgo.Collection
)

func init() {
	url := fmt.Sprintf("%v:%v", common.Config.Db.Host, common.Config.Db.Port)
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

func GetMangaAll() ([]Manga, error) {
	var manga []Manga
	return manga, mangaCollection.
		Find(bson.M{}).
		Sort("url").
		Select(bson.M{
			"chapters": 0,
		}).
		All(&manga)
}

func GetMangaAllMin() ([]MangaMin, error) {
	var manga []MangaMin
	return manga, mangaCollection.
		Find(bson.M{}).
		Sort("url").
		Select(bson.M{
			"chapters": 0,
			"titles":   0,
			"srcUrl":   0,
		}).
		All(&manga)
}

func GetRanobeAll() ([]Ranobe, error) {
	var ranobe []Ranobe
	ranobe = []Ranobe{}
	return ranobe, nil
}

func GetManga(mangaUrl string) (Manga, error) {
	findSelector := bson.M{
		"url": mangaUrl,
	}

	manga, ok := mangaCache.Find(mangaUrl)
	if !ok {
		if err := mangaCollection.Find(findSelector).One(&manga); err != nil {
			return manga, err
		}
		mangaCache.Add(manga)
	}

	return manga, nil
}

func GetMangaImages(mangaUrl string, chapter int) ([]string, error) {
	findSelector := bson.M{
		"manga":  mangaUrl,
		"number": chapter,
	}
	selectSelector := bson.M{
		"image": 1,
	}
	var structImages []MangaImage

	images, ok := imageCache.Find(mangaUrl, chapter)
	if !ok {
		err := mangaImagesCollection.
			Find(findSelector).
			Sort("image").
			Select(selectSelector).
			All(&structImages)
		if err != nil {
			return images, err
		}
		for _, image := range structImages {
			images = append(images, image.Image)
		}
		imageCache.Add(mangaUrl, chapter, images)
	}

	return images, nil
}

func AddManga(manga Manga) error {
	findSelector := bson.M{
		"url": manga.Url,
	}

	if num, err := mangaCollection.Find(findSelector).Count(); err != nil {
		return err
	} else if num != 0 {
		return fmt.Errorf("manga is already added")
	}

	if err := mangaCollection.Insert(&manga); err != nil {
		return err
	}
	mangaCache.Add(manga)
	return nil
}

func RemoveManga(mangaUrl string) error {
	findSelector := bson.M{
		"url": mangaUrl,
	}

	mangaCache.Remove(mangaUrl)
	if err := mangaCollection.Remove(findSelector); err != nil {
		return err
	}

	imageCache.Remove(mangaUrl)
	return mangaImagesCollection.Remove(findSelector)
}

func AddMangaChapter(name string, chapter MangaChapter, images []string) error {
	var result Manga
	findSelector := bson.M{
		"url": name,
	}
	updChaptersSelector := bson.M{
		"$push": bson.M{
			"chapters": chapter,
		},
	}
	updDateSelector := bson.M{
		"$set": bson.M{
			"upddate": time.Now(),
			"size":    result.Size + 1,
		},
	}

	mangaCache.Remove(name)
	if err := mangaCollection.Find(findSelector).One(&result); err != nil {
		return err
	}
	if err := mangaCollection.Update(findSelector, updChaptersSelector); err != nil {
		return err
	}
	if err := mangaCollection.Update(findSelector, updDateSelector); err != nil {
		return err
	}

	imageCache.Add(name, chapter.Number, images)
	for _, image := range images {
		imgObj := MangaImage{
			Manga:  name,
			Number: chapter.Number,
			Image:  image,
		}
		if err := mangaImagesCollection.Insert(imgObj); err != nil {
			return err
		}
	}

	return nil
}

func AddMangaTitle(mangaUrl, titleName string) error {
	findSelector := bson.M{
		"url": mangaUrl,
	}
	updSelector := bson.M{
		"$push": bson.M{
			"titles": titleName,
		},
	}
	mangaCache.Remove(mangaUrl)
	return mangaCollection.Update(findSelector, updSelector)
}

func RemoveMangaTitle(mangaUrl, titleName string) error {
	findSelector := bson.M{
		"url": mangaUrl,
	}
	updSelector := bson.M{
		"$pull": bson.M{
			"titles": titleName,
		},
	}
	mangaCache.Remove(mangaUrl)
	return mangaCollection.Update(findSelector, updSelector)
}

func RemoveMangaChapter(mangaUrl string, chapNumber int) error {
	findSelector := bson.M{
		"url": mangaUrl,
	}
	updSelector := bson.M{
		"$pull": bson.M{
			"chapters": bson.M{
				"number": chapNumber,
			},
		},
	}
	rmSelector := bson.M{"manga": mangaUrl, "number": chapNumber}

	mangaCache.Remove(mangaUrl)
	imageCache.Remove(mangaUrl)

	if err := mangaCollection.Update(findSelector, updSelector); err != nil {
		return err
	}
	return mangaImagesCollection.Remove(rmSelector)
}

func ChangeMangaName(mangaUrl, name string) error {
	findSelector := bson.M{
		"url": mangaUrl,
	}
	updSelector := bson.M{
		"$set": bson.M{
			"name": name,
		},
	}

	mangaCache.Remove(mangaUrl)
	return mangaCollection.Update(findSelector, updSelector)
}

func ChangeMangaChapName(mangaUrl string, chapNumber int, chapName string) error {
	findSelector := bson.M{
		"url":             mangaUrl,
		"chapters.number": chapNumber,
	}
	updSelector := bson.M{
		"$set": bson.M{
			"chapters.$.name": chapName,
		},
	}

	mangaCache.Remove(mangaUrl)
	return mangaCollection.Update(findSelector, updSelector)
}
