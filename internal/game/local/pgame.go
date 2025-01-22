package local

import (
	"github.com/joakim-ribier/pong/internal/drawer"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/pkg"
)

type LocalPGame struct {
	drawer *drawer.GameDrawer
}

func NewPGame(debug bool, version string) *LocalPGame {
	return &LocalPGame{
		drawer: drawer.NewDrawerGame(
			pkg.NewGame(pkg.LocalMode, debug),
			func(network.Message) {}, func() {},
			version),
	}
}

// Title returns the console title
func (pg *LocalPGame) Title() string {
	return pg.drawer.Game.Title.Text
}

// Drawer returns the drawer that builds the game
func (pg *LocalPGame) Drawer() *drawer.GameDrawer {
	return pg.drawer
}
