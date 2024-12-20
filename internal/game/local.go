package game

import (
	"github.com/joakim-ribier/pong/internal/drawer"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/pkg"
)

type LocalPongGame struct {
	drawer *drawer.GameDrawer
}

func NewLocalPongGame(debug bool, version string) *LocalPongGame {
	return &LocalPongGame{
		drawer: drawer.NewDrawerGame(
			pkg.NewGame(pkg.LocalMode, debug),
			func(network.Message) {}, func() bool { return true },
			version),
	}
}

// Title returns the console title
func (pg *LocalPongGame) Title() string {
	return pg.drawer.Game.Title.Text
}

// Drawer returns the drawer that builds the game
func (pg *LocalPongGame) Drawer() *drawer.GameDrawer {
	return pg.drawer
}
