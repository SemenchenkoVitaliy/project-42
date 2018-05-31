package mongo

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/SemenchenkoVitaliy/project-42/utils"
)

type DatabaseUsers struct {
	users *mgo.Collection
}

func NewDatabaseUsers() (db *DatabaseUsers) {
	return &DatabaseUsers{}
}

func (db *DatabaseUsers) Connect(user, password, dbName, ip string, port int) {
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
	db.users = session.DB(dbName).C("users")
}

type User struct {
	Email     string
	Name      string
	Rights    int
	Sessions  []string
	Bookmarks []string
}

func (db *DatabaseUsers) AddUser(email, name, session string) (err error) {
	var user User
	err = db.users.
		Find(bson.M{
			"email": email,
		}).
		One(&user)
	if err != nil {
		if err.Error() == "not found" {
			return db.users.
				Insert(&User{
					Email:    email,
					Name:     name,
					Sessions: []string{session},
				})
		} else {
			return err
		}
	} else {
		return db.users.
			Update(bson.M{
				"email": email,
			}, bson.M{
				"$push": bson.M{
					"sessions": session,
				},
			})
	}
}

func (db *DatabaseUsers) RemoveUserSession(session string) (err error) {
	return db.users.
		Update(bson.M{
			"sessions": bson.M{
				"$in": []string{session},
			},
		}, bson.M{
			"$pull": bson.M{
				"sessions": session,
			},
		})
}

func (db *DatabaseUsers) FindUser(session string) (user User, err error) {
	err = db.users.
		Find(bson.M{
			"sessions": bson.M{
				"$in": []string{session},
			},
		}).
		One(&user)
	return user, err
}
