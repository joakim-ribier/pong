package drawer

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/pong/pkg"
)

const IMPRESSIONS_MAX = 120
const SPEED_RATIO = 1.05

type BallDrawer struct {
	ball     *pkg.Ball
	playerL  pkg.Player
	playerR  pkg.Player
	debug    bool
	isClient bool
}

func NewBallDrawer(game pkg.Game) *BallDrawer {
	return &BallDrawer{
		ball:     game.Ball,
		playerL:  *game.PlayerL,
		playerR:  *game.PlayerR,
		debug:    game.Debug,
		isClient: game.IsRemoteClient(),
	}
}

func (b *BallDrawer) Draw(screen *ebiten.Image) {
	DrawImage(screen, b.ball.Image, b.ball.Position)

	// display ball impressions for debug
	for _, imp := range b.ball.Impressions {
		DrawImage(screen, &imp.Image, imp.Position)
	}
}

func (b *BallDrawer) Update(game *pkg.Game, screen pkg.Screen, demo bool) {
	borderMarginY := float32(15)

	b.ball.X += b.ball.XSpeed
	b.ball.Y += b.ball.YSpeed

	if b.ball.Y+float32(b.ball.Height) >= float32(screen.YTop)-borderMarginY {
		b.ball.YSpeed = -b.ball.YSpeed
		b.ball.Y = float32(screen.YTop) - float32(b.ball.Height) - borderMarginY
	} else if b.ball.Y <= float32(screen.YBottom)+borderMarginY {
		b.ball.YSpeed = -b.ball.YSpeed
		b.ball.Y = float32(screen.YBottom) + borderMarginY
	}

	if b.playerL.Hit(*b.ball) {
		b.ball.XSpeed = -b.ball.XSpeed * genericsutil.OrElse[float32](SPEED_RATIO, func() bool { return !demo }, 1)
		b.ball.X = b.playerL.Paddle.X + float32(b.playerL.Paddle.Width)
		game.Hit()
	} else if b.playerR.Hit(*b.ball) {
		b.ball.XSpeed = -b.ball.XSpeed * genericsutil.OrElse[float32](SPEED_RATIO, func() bool { return !demo }, 1)
		b.ball.X = b.playerR.Paddle.X - float32(b.playerR.Paddle.Width/2) - float32(b.ball.Width/2)
		game.Hit()
	}

	game.SetXSpeed(b.ball.XSpeed)

	if b.debug {
		// stack the impressions for debug
		b.ball.Impressions = append(b.ball.Impressions, pkg.Impression{Position: b.ball.Position, Image: *b.ball.Image})
		if nb := len(b.ball.Impressions); nb > IMPRESSIONS_MAX {
			b.ball.Impressions = b.ball.Impressions[nb-IMPRESSIONS_MAX : IMPRESSIONS_MAX]
		}
	} else {
		b.ball.Impressions = nil
	}
}
