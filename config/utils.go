package config

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand"

	"github.com/atotto/clipboard"
)

func CopyToClipboard(text string) {
	err := clipboard.WriteAll(text)
	if err != nil {
		log.Println("Can't copy ip address to clipboard")
	}
}

func FourBytesToInt32(buff []byte) int32 {
	return int32(binary.BigEndian.Uint32(buff))
}

func Uint64ToBytes(value uint64) []byte {
	tempBuff := make([]byte, 8)
	binary.BigEndian.PutUint64(tempBuff, value)
	return tempBuff
}

func Uint32ToBytes(value uint32) []byte {
	tempBuff := make([]byte, 4)
	binary.BigEndian.PutUint32(tempBuff, value)
	return tempBuff
}

func Uint16ToBytes(value uint16) []byte {
	tempBuff := make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuff, value)
	return tempBuff
}

func Int32ToBytes(value int32) []byte {
	var tempBuff bytes.Buffer
	binary.Write(&tempBuff, binary.BigEndian, value)
	return tempBuff.Bytes()
}

func Generate20ByteRandomString() string {
	peerIdLength := 20 // this is always a constant
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := ""
	for i := 0; i < peerIdLength; i++ {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}
