package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/pong/pkg"
	"github.com/joakim-ribier/pong/pkg/drawer"
)

func main() {
	debug := flag.Bool("debug", false, "help message for flag n")
	flag.Parse()

	drawer := drawer.NewDrawerGame(pkg.NewGame(*debug))

	ebiten.SetWindowSize(drawer.Game.Screen.Width, drawer.Game.Screen.Height)
	ebiten.SetWindowTitle(drawer.Game.Title.Text)

	if err := ebiten.RunGame(drawer); err != nil {
		log.Fatal(err)
	}
}
