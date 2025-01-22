package pkg

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/pong/pkg/resources"
)

type Ball struct {
	Position
	XSpeed      float32
	YSpeed      float32
	Width       int
	Height      int
	Image       *ebiten.Image
	Impressions []Impression

	UpdateBall chan Position
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
		Image:       GetImg(resources.BallWhitex16),
		Impressions: nil,
		UpdateBall:  make(chan Position, 256),
	}
}
