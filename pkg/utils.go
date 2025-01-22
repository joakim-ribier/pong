package pkg

import (
	"bytes"
	"image"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

func GetImg(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}

func ToUDPAddrUnsafe(addr string) *net.UDPAddr {
	addrT := strings.Split(addr, ":")

	ip := addrT[0]
	port := 0
	if p, err := strconv.ParseInt(addrT[1], 10, 64); err == nil {
		port = int(p)
	}

	return &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
		Zone: "",
	}
}
