package drawer

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/pong/pkg"
)

type PlayersDrawer struct {
	game        pkg.Game
	PlayerLeft  pkg.Player
	PlayerRight pkg.Player
}

func NewPlayerDrawer(game pkg.Game) *PlayersDrawer {
	return &PlayersDrawer{
		game:        game,
		PlayerLeft:  *game.PlayerL,
		PlayerRight: *game.PlayerR}
}

func (p *PlayersDrawer) Draw(screen *ebiten.Image) {
	NewPaddleDrawer(p.PlayerLeft).Draw(screen)
	NewPaddleDrawer(p.PlayerRight).Draw(screen)
}

func (p *PlayersDrawer) Update(screen pkg.Screen) {
	p.update(p.PlayerLeft, screen)
	p.update(p.PlayerRight, screen)
}

func (p *PlayersDrawer) UpdateLeft(screen pkg.Screen) pkg.State {
	return p.update(p.PlayerLeft, screen)
}

func (p *PlayersDrawer) UpdateRight(screen pkg.Screen) pkg.State {
	return p.update(p.PlayerRight, screen)
}

func (p *PlayersDrawer) update(player pkg.Player, screen pkg.Screen) pkg.State {
	NewPaddleDrawer(player).Update(screen)

	if player.Side == pkg.PlayerLeft && p.game.Ball.X < player.Paddle.X {
		return pkg.PlayerLLostBall
	} else if player.Side == pkg.PlayerRight && p.game.Ball.X > player.Paddle.X {
		return pkg.PlayerRLostBall
	}

	return p.game.CurrentState
}
