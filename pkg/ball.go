package pkg

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Ball struct {
	Position
	XSpeed      float32
	YSpeed      float32
	Width       int
	Height      int
	Image       *ebiten.Image
	Impressions []Impression
}

type Impression struct {
	Position
	Image ebiten.Image
}

func NewBall(w, h int, position Position) *Ball {
	return &Ball{
		Position:    position,
		XSpeed:      5,
		YSpeed:      5,
		Width:       w,
		Height:      h,
		Image:       GetImg("ball-white", w),
		Impressions: nil,
	}
}
