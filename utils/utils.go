// pakage utils provides config loaded from config file and command options and
// logging functions
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GetUUID() string {
	u := [16]byte{}
	rand.Reader.Read(u[:])

	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

func ProcessCovers(covers []string, url string) []string {
	if len(covers) == 0 {
		return []string{"/static/no-cover.png"}
	}

	images := make([]string, len(covers))

	for index, image := range covers {
		images[index] = fmt.Sprintf(
			"http://img.%v/images/mangaCovers/%v/%v",
			Config.PublicUrl,
			url,
			image,
		)
	}
	return images
}

func ProcessPages(pages []string, name, chapter string) []string {
	images := make([]string, len(pages))

	for index, image := range pages {
		images[index] = fmt.Sprintf(
			"http://img.%v/images/manga/%v/%v/%v",
			Config.PublicUrl,
			name,
			chapter,
			image,
		)
	}
	return images
}
