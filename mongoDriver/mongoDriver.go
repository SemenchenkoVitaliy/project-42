package mongoDriver

import (
	"fmt"
	"os"
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
	session, err := mgo.Dial(fmt.Sprintf("%v:%v", common.Config.Db.Host, common.Config.Db.Port))
	if err != nil {
		common.CreateLog(err, "start MongoDB session")
		os.Exit(1)
	}
	if err := session.DB(common.Config.Db.DbName).Login(common.Config.Db.User, common.Config.Db.Password); err != nil {
		common.CreateLog(err, "authenticate MongoDB session")
		os.Exit(1)
	}

	session.SetMode(mgo.Monotonic, true)

	mangaCollection = session.DB("gotest").C("mangalist")
	mangaImagesCollection = session.DB("gotest").C("mangaimages")
}

func GetMangaAll() []Manga {
	var manga []Manga

	mangaCollection.Find(bson.M{}).Sort("url").All(&manga)

	return manga
}

func GetRanobeAll() []Ranobe {
	var ranobe []Ranobe

	ranobe = []Ranobe{}

	return ranobe
}

func GetManga(mangaUrl string) (Manga, error) {
	var manga Manga

	err := mangaCollection.Find(bson.M{"url": mangaUrl}).One(&manga)

	return manga, err
}

func GetMangaImages(mangaUrl string, chapter int) []string {
	var structImages []MangaImage
	stringImages := []string{}

	mangaImagesCollection.Find(bson.M{"manga": mangaUrl, "number": chapter}).Sort("image").Select(bson.M{"image": 1}).All(&structImages)

	for _, image := range structImages {
		stringImages = append(stringImages, image.Image)
	}

	return stringImages
}

func AddManga(manga Manga) error {
	num, err := mangaCollection.Find(bson.M{"url": manga.Url}).Count()
	if err != nil {
		return err
	}

	if num == 0 {
		mangaCollection.Insert(&manga)
		return nil
	} else {
		return fmt.Errorf("manga is already added")
	}
}

func RemoveManga(mangaUrl string) error {
	err := mangaCollection.Remove(bson.M{"url": mangaUrl})
	if err != nil {
		return err
	}
	err = mangaImagesCollection.Remove(bson.M{"manga": mangaUrl})
	return err
}

func AddMangaChapter(name string, chapter MangaChapter, images []string) error {
	var result Manga
	findQuery := bson.M{"url": name}

	err := mangaCollection.Find(findQuery).One(&result)
	if err != nil {
		return err
	}

	err = mangaCollection.Update(findQuery, bson.M{"$push": bson.M{"chapters": chapter}})
	if err != nil {
		return err
	}

	err = mangaCollection.Update(findQuery, bson.M{"$set": bson.M{"upddate": time.Now(), "size": result.Size + 1}})
	if err != nil {
		return err
	}

	for _, image := range images {
		err = mangaImagesCollection.Insert(MangaImage{Manga: name, Number: chapter.Number, Image: image})
		if err != nil {
			return err
		}
	}

	return nil
}

func AddMangaTitle(mangaUrl, titleName string) error {
	return mangaCollection.Update(bson.M{"url": mangaUrl}, bson.M{"$push": bson.M{"titles": titleName}})
}

func RemoveMangaTitle(mangaUrl, titleName string) error {
	return mangaCollection.Update(bson.M{"url": mangaUrl}, bson.M{"$pull": bson.M{"titles": titleName}})
}

func RemoveMangaChapter(mangaUrl string, chapNumber int) error {
	err := mangaCollection.Update(bson.M{"url": mangaUrl}, bson.M{"$pull": bson.M{"chapters": bson.M{"number": chapNumber}}})
	if err != nil {
		return err
	}

	err = mangaImagesCollection.Remove(bson.M{"manga": mangaUrl, "number": chapNumber})
	return err
}

func ChangeMangaName(mangaUrl, name string) error {
	return mangaCollection.Update(bson.M{"url": mangaUrl}, bson.M{"$set": bson.M{"name": name}})
}

func ChangeMangaChapName(mangaUrl string, chapNumber int, chapName string) error {
	return mangaCollection.Update(bson.M{"url": mangaUrl, "chapters.number": chapNumber}, bson.M{"$set": bson.M{"chapters.$.name": chapName}})
}
