package fileserver

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/cache"
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

var fileCache = cache.NewCache(1000, 60)

func root(w http.ResponseWriter, r *http.Request) {
	path := utils.Config.SrcDir + r.URL.Path
	if strings.HasSuffix(r.URL.Path, "/") {
		netutils.ForbiddenError(w, nil, "403 - access forbidden")
		return
	}
	var err error
	data, ok := fileCache.Get(path)
	if !ok {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			netutils.NotFoundError(w, nil, "404 - not found")
			return
		}
		fileCache.Add(path, data)
	}

	w.Write(data.([]byte))
}

func rootHashed(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		netutils.ForbiddenError(w, nil, "403 - access forbidden")
		return
	}

	var err error
	data, ok := fileCache.Get(r.URL.Path)
	if !ok {
		h := sha256.New()
		h.Write([]byte(r.URL.Path))
		path := utils.Config.SrcDir + "/" + base64.URLEncoding.EncodeToString(h.Sum(nil))
		data, err = ioutil.ReadFile(path)
		if err != nil {
			netutils.NotFoundError(w, nil, "404 - not found")
			return
		}
		fileCache.Add(r.URL.Path, data)
	}

	w.Write(data.([]byte))
}
