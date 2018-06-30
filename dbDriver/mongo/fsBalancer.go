package mongo

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/SemenchenkoVitaliy/project-42/utils"
)

type DatabaseFS struct {
	fs *mgo.Collection
}

func NewDatabaseFS() (db *DatabaseFS) {
	return &DatabaseFS{}
}

func (db *DatabaseFS) Connect(user, password, dbName, ip string, port int) {
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
	db.fs = session.DB(dbName).C("fs")
}

type entry struct {
	Url     string
	Numbers []int
}

func (db *DatabaseFS) AddEntry(url string, sNumbers []int) (err error) {
	db.fs.Remove(bson.M{
		"url": url,
	})
	return db.fs.
		Insert(&entry{
			Url:     url,
			Numbers: sNumbers,
		})
}

func (db *DatabaseFS) RemoveEntry(url string) (err error) {
	return db.fs.
		Remove(bson.M{
			"url": url,
		})
}

func (db *DatabaseFS) FindEntry(url string) (sNumbers []int, err error) {
	var e entry
	err = db.fs.
		Find(bson.M{
			"url": url,
		}).
		One(&e)
	return e.Numbers, err
}
