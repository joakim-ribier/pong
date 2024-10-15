package pkg

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func GetImg(filename string, size int) *ebiten.Image {
	data, err := os.ReadFile(fmt.Sprintf("../../resources/%s-%d.png", filename, size))
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}
