package parse

import (
	"fmt"
	"log"
	"os"
)

func (t Torrent) ReassemblePieces() {
	folderDestination := t.Info.Name
	fileName := t.Info.Name

	fileFd, err := os.Create(fmt.Sprintf("./%s/%s", folderDestination, fileName))
	if err != nil {
		log.Print("Error creating the download file")
		log.Fatal(err)
	}
	defer fileFd.Close()

	for _,piece := range t.PieceHashes {
		pieceFile := fmt.Sprintf("./%s/%x", folderDestination, piece)
		buff, err := os.ReadFile(pieceFile)
		if err != nil {
      log.Print(err)
			log.Printf("Error writing %x piece\n", piece)
		}
		fileFd.Write(buff)
		os.Remove(pieceFile)
	}
}
