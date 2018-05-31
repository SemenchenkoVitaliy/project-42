package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/dbDriver/mongo"
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

var (
	db      = dbDriver.NewDatabase()
	dbUsers = dbDriver.NewDatabaseUsers()

	publicUrl string
	templates *template.Template

	googleOauthConfig *oauth2.Config
	sessions          map[string]string = make(map[string]string)
)

func loadTemplates() {
	templates = template.Must(template.ParseGlob("./HTML/*.gohtml"))
}

func loadOauthConfig(id, secret string) {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://" + publicUrl + "/googleCallback",
		ClientID:     id,
		ClientSecret: secret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func checkUser(w http.ResponseWriter, r *http.Request) (user dbDriver.User, err error) {
	c, err := r.Cookie("session")
	if err != nil {
		return user, err
	}
	return dbUsers.FindUser(c.Value)
}

func index(w http.ResponseWriter, r *http.Request) {
	mangaUrls, err := db.GetMangaUrls(0, 10, "upddate")
	if err != nil {
		netutils.InternalError(w, err, "Get top manga urls")
		return
	}

	manga, err := db.GetMangaMultiple(mangaUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top manga")
		return
	}

	for index, item := range manga {
		manga[index].Covers = utils.ProcessCovers(item.Covers, item.Url)
	}

	ranobeUrls, err := db.GetRanobeUrls(0, 10, "upddate")
	if err != nil {
		netutils.InternalError(w, err, "Get top ranobe urls")
		return
	}

	ranobe, err := db.GetRanobeMultiple(ranobeUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top ranobe")
		return
	}

	user, _ := checkUser(w, r)

	data := struct {
		Manga    []dbDriver.Product
		Ranobe   []dbDriver.Product
		UserName string
		Rights   int
	}{
		Manga:    manga,
		Ranobe:   ranobe,
		UserName: user.Name,
		Rights:   user.Rights,
	}

	if err = templates.ExecuteTemplate(w, "index", data); err != nil {
		netutils.InternalError(w, err, "Execute template index")
		return
	}
}

func manga(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	mangaUrls, err := db.GetMangaUrls(page*10, 10, "name")
	if err != nil {
		netutils.InternalError(w, err, "Get top manga urls")
		return
	}

	data, err := db.GetMangaMultiple(mangaUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top manga")
		return
	}

	for index, item := range data {
		data[index].Covers = utils.ProcessCovers(item.Covers, item.Url)
	}

	if err = templates.ExecuteTemplate(w, "mangaAll", data); err != nil {
		netutils.InternalError(w, err, "Execute template mangaAll")
		return
	}
}

func mangaInfo(w http.ResponseWriter, r *http.Request) {
	data, err := db.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		netutils.NotFoundError(w, nil, "No such manga: "+mux.Vars(r)["name"])
		return
	}

	data.Covers = utils.ProcessCovers(data.Covers, data.Url)

	if err = templates.ExecuteTemplate(w, "mangaInfo", data); err != nil {
		netutils.InternalError(w, err, "Execute template mangaInfo")
		return
	}
}

func mangaChapter(w http.ResponseWriter, r *http.Request) {
	chapNumber, err := strconv.Atoi(mux.Vars(r)["chapter"])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	data, err := db.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		netutils.NotFoundError(w, nil, "No such manga: "+mux.Vars(r)["name"])
		return
	}

	images := utils.ProcessPages(
		data.Chapters[chapNumber].Pages,
		mux.Vars(r)["name"],
		mux.Vars(r)["chapter"])

	result := struct {
		Manga          dbDriver.Product
		Images         []string
		CurrentChapter int
		PublicUrl      string
	}{
		Manga:          data,
		Images:         images,
		CurrentChapter: chapNumber,
		PublicUrl:      publicUrl,
	}

	if err = templates.ExecuteTemplate(w, "mangaChapter", result); err != nil {
		netutils.InternalError(w, err, "Execute template mangaChapter")
		return
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	user, err := checkUser(w, r)
	if err != nil || user.Rights != 1 {
		netutils.UnauthorizedError(w, err, "Unauthorized user")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	mangaUrls, err := db.GetMangaUrls(page*10, 10, "name")
	if err != nil {
		netutils.InternalError(w, err, "Get top 10 manga urls")
		return
	}

	manga, err := db.GetMangaMultiple(mangaUrls)
	for index, item := range manga {
		manga[index].Covers = utils.ProcessCovers(item.Covers, item.Url)
	}

	ranobeUrls, err := db.GetRanobeUrls(page*10, 10, "name")
	if err != nil {
		netutils.InternalError(w, err, "Get top 10 ranobe urls")
		return
	}

	ranobe, err := db.GetRanobeMultiple(ranobeUrls)
	if err != nil {
		netutils.InternalError(w, err, "Get top 10 ranobe")
		return
	}

	data := struct {
		Manga     []dbDriver.Product
		Ranobe    []dbDriver.Product
		PublicUrl string
		UserName  string
	}{
		Manga:     manga,
		Ranobe:    ranobe,
		PublicUrl: publicUrl,
		UserName:  user.Name,
	}

	if err = templates.ExecuteTemplate(w, "admin", data); err != nil {
		netutils.InternalError(w, err, "Execute template admin")
		return
	}
}

func adminMangaInfo(w http.ResponseWriter, r *http.Request) {
	user, err := checkUser(w, r)
	if err != nil || user.Rights != 1 {
		netutils.UnauthorizedError(w, err, "Unauthorized user")
		return
	}

	manga, err := db.GetMangaSingle(mux.Vars(r)["name"])
	if err != nil {
		netutils.NotFoundError(w, err, "Get manga "+mux.Vars(r)["name"])
		return
	}

	data := struct {
		Manga     dbDriver.Product
		PublicUrl string
		UserName  string
	}{
		Manga:     manga,
		PublicUrl: publicUrl,
		UserName:  user.Name,
	}

	if err = templates.ExecuteTemplate(w, "adminMangaInfo", data); err != nil {
		netutils.InternalError(w, err, "Execute template adminMangaInfo")
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	src := [32]byte{}
	rand.Reader.Read(src[:])
	buf := make([]byte, 64)
	hex.Encode(buf, src[:])

	state := string(buf)
	session := utils.GetUUID()
	sessions[session] = state

	url := googleOauthConfig.AuthCodeURL(state)
	http.SetCookie(w, &http.Cookie{Name: "session", Value: session, Path: "/"})
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("session")
	dbUsers.RemoveUserSession(c.Value)
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func googleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	c, err := r.Cookie("session")
	stateExpected := sessions[c.Value]
	delete(sessions, c.Value)

	if err != nil || state != stateExpected {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", stateExpected, state)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Code exchange failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	response, _ := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer response.Body.Close()
	contents, _ := ioutil.ReadAll(response.Body)

	var Content struct {
		Email string
		Name  string
	}

	json.Unmarshal(contents, &Content)

	dbUsers.AddUser(Content.Email, Content.Name, c.Value)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}
