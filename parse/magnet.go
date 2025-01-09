package parse

import (
	"log"

	"github.com/kyoto44/rain/magnet"
)

func ParseMagnetLink(magnetLink string) *magnet.Magnet {
	mag, err := magnet.New(magnetLink)
	if err != nil {
		log.Fatalln("Can't parse magnet link")
	}
	return mag
}
