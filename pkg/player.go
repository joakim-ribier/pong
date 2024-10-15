package pkg

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player is a player with a paddle and options
type Player struct {
	*PlayerState

	Name    string
	Side    PlayerSide
	Paddle  *Paddle
	Options Options
}

// Player is a player with a paddle and options
type PlayerState struct {
	Score int
	Win   bool
}

// Options is an options (shortcuts, color, etc...)
type Options struct {
	Color color.Color
	Up    ebiten.Key
	Down  ebiten.Key
}

func NewPlayer(name string, side PlayerSide, p *Paddle, options Options) *Player {
	return &Player{
		Name:        name,
		Side:        side,
		Paddle:      p,
		Options:     options,
		PlayerState: &PlayerState{0, false},
	}
}

func (p *Player) Mark() {
	p.Score++
	p.Win = true
}

func (p *Player) Hit(ball Ball) bool {
	behindPaddle := false
	if p.Side == PlayerLeft {
		behindPaddle = ball.X <= p.Paddle.X+float32(p.Paddle.Width)
	} else if p.Side == PlayerRight {
		behindPaddle = ball.X+float32(ball.Height) >= p.Paddle.X
	}
	return behindPaddle &&
		ball.Y >= p.Paddle.Y && ball.Y <= p.Paddle.Y+float32(p.Paddle.Height)
}

// PlayerSide is an enum that represents the position of the player on the screen (LEFT or RIGHT)
type PlayerSide int

const (
	PlayerLeft PlayerSide = iota
	PlayerRight
)

func (p PlayerSide) String() string {
	switch p {
	case PlayerLeft:
		return "PlayerLeft"
	case PlayerRight:
		return "PlayerRight"
	default:
		return "Unknown"
	}
}
