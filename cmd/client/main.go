package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/pong/internal/game"
	"github.com/joakim-ribier/pong/pkg"
)

func main() {
	debug := flag.Bool("debug", false, "define true to display more debug details...")
	server := flag.String("server", "", "start the server and set its address")
	client := flag.String("client", "", "start a new client and define the server address")
	flag.Parse()

	var pongGame game.PongGame

	if *server != "" {
		pongGame = game.NewRemotePongGame(*debug, pkg.RemoteServerMode, *server)
	} else if *client != "" {
		pongGame = game.NewRemotePongGame(*debug, pkg.RemoteClientMode, *client)
	} else {
		pongGame = game.NewLocalPongGame(*debug)
	}

	ebiten.SetWindowTitle(pongGame.Title())
	ebiten.SetWindowSize(
		pongGame.Drawer().Game.Screen.Width,
		pongGame.Drawer().Game.Screen.Height,
	)

	if err := ebiten.RunGame(pongGame.Drawer()); err != nil {
		log.Fatal(err)
	}
}
