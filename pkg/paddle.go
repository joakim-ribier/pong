package pkg

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Paddle struct {
	Position
	Speed          float32
	Width          int
	Height         int
	Image          *ebiten.Image
	CurrentPressed ebiten.Key
}

func NewPaddle(w, h int, position Position) *Paddle {
	return &Paddle{
		Position:       position,
		Speed:          10,
		Width:          w,
		Height:         h,
		Image:          ebiten.NewImage(w, h),
		CurrentPressed: -1,
	}
}
