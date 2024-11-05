package game

import (
	"github.com/joakim-ribier/pong/internal/drawer"
)

type PongGame interface {
	// Title returns the console title
	Title() string
	// Drawer returns the drawer that builds the game
	Drawer() *drawer.GameDrawer
}
