package config

import (
	"log"

	"github.com/atotto/clipboard"
)

func CopyToClipboard(text string) {
  err:=   clipboard.WriteAll(text)
  if err!= nil{
    log.Println("Can't copy ip address to clipboard")
  }
}
