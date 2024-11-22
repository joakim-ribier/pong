package drawer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/joakim-ribier/pong/pkg"
)

func getOptions(position pkg.Position, color color.Color) *text.DrawOptions {
	pOpts := &text.DrawOptions{}
	pOpts.GeoM.Translate(float64(position.X), float64(position.Y))
	pOpts.ColorScale.ScaleWithColor(color)

	return pOpts
}

func getRectangle(w, h int, position pkg.Position, color color.Color) (*ebiten.Image, *ebiten.DrawImageOptions) {
	img := ebiten.NewImage(w, h)

	pOpts := &ebiten.DrawImageOptions{}
	pOpts.GeoM.Translate(float64(position.X), float64(position.Y))

	img.Fill(color)

	return img, pOpts
}

// GetXCenterPos gets the x position to be center from {w} value
func GetXCenterPos(w int, text string, fontSize int) int {
	return (w - len(text)*fontSize) / 2
}

// DrawText draws text at the specific position
func DrawText(screen *ebiten.Image, str string, face text.Face, color color.Color, position pkg.Position) {
	text.Draw(screen, str, face, getOptions(position, color))
}

// DrawRectangle draws a rectangle at the specific position
func DrawRectangle(screen *ebiten.Image, w, h int, position pkg.Position, color color.Color) {
	img, options := getRectangle(w, h, position, color)
	screen.DrawImage(img, options)
}

// DrawImage draws the {img} on the {screen} at the specific {position}
func DrawImage(screen *ebiten.Image, img *ebiten.Image, position pkg.Position) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(float64(position.X), float64(position.Y))
	screen.DrawImage(img, options)
}

// GetSize compute the size of the {text} field
func GetSize(text string, fontSize int) int {
	return len(text) * fontSize
}
