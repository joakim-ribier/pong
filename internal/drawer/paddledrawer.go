package drawer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/joakim-ribier/pong/pkg"
)

type PaddleDrawer struct {
	paddle  *pkg.Paddle
	color   color.Color
	side    pkg.PlayerSide
	options pkg.Options
}

func NewPaddleDrawer(player pkg.Player) *PaddleDrawer {
	return &PaddleDrawer{
		paddle:  player.Paddle,
		color:   player.Options.Color,
		side:    player.Side,
		options: player.Options,
	}
}

func (p *PaddleDrawer) Draw(screen *ebiten.Image) {
	pOpts := &ebiten.DrawImageOptions{}
	pOpts.GeoM.Translate(float64(p.paddle.X), float64(p.paddle.Y))

	p.paddle.Image.Fill(p.color)

	screen.DrawImage(p.paddle.Image, pOpts)
}

func (p *PaddleDrawer) Update(screen pkg.Screen, updateY bool) {
	borderMarginX := float32(25)
	borderMarginY := float32(15)

	if p.side == pkg.PlayerLeft {
		p.paddle.X = screen.XLeft + borderMarginX
	} else if p.side == pkg.PlayerRight {
		p.paddle.X = screen.XRight - float32(p.paddle.Width) - borderMarginX
	}

	if updateY {
		p.updateY()
	}

	if p.paddle.Y <= float32(screen.YBottom)+borderMarginY {
		p.paddle.Y = screen.YBottom + borderMarginY
	} else if p.paddle.Y+float32(p.paddle.Height) >= float32(screen.YTop)-borderMarginY {
		p.paddle.Y = screen.YTop - float32(p.paddle.Height) - borderMarginY
	}
}

func (p *PaddleDrawer) updateY() {
	if inpututil.IsKeyJustPressed(p.options.Up) {
		p.paddle.CurrentPressed = p.options.Up
	} else if inpututil.IsKeyJustReleased(p.options.Up) {
		p.paddle.CurrentPressed = -1
	}

	if inpututil.IsKeyJustPressed(p.options.Down) {
		p.paddle.CurrentPressed = p.options.Down
	} else if inpututil.IsKeyJustReleased(p.options.Down) {
		p.paddle.CurrentPressed = -1
	}

	if p.paddle.CurrentPressed == p.options.Up {
		p.paddle.Y -= p.paddle.Speed
	}

	if p.paddle.CurrentPressed == p.options.Down {
		p.paddle.Y += p.paddle.Speed
	}
}
