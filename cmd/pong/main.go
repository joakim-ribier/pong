package main

import (
	"flag"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/pong/internal/game"
	"github.com/joakim-ribier/pong/internal/game/local"
	"github.com/joakim-ribier/pong/internal/game/online"
	"github.com/joakim-ribier/pong/pkg"
	"github.com/joakim-ribier/pong/pkg/resources"
)

func main() {
	debug := flag.Bool("debug", false, "enable the 2-D engine [debug] mode")
	server := flag.String("server", "", "start a server [--server 0.0.0:3000] to host the game")
	client := flag.String("client", "", "start a client [--client 0.0.0:3000] to connect to the server")
	verbose := flag.Bool("verbose", false, "enable the [verbose] mode to display logs")

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if !*verbose {
		log.SetOutput(io.Discard)
	}

	pGame := genericsutil.When[*onlineMode, game.PGame](
		parseOnlineModeParam(*server, *client), func(p *onlineMode) bool { return p != nil },
		func(om *onlineMode) game.PGame {
			return online.NewPGame(*debug, om.gameMode(), om.addr, resources.Version)
		},
		func() game.PGame { return local.NewPGame(*debug, resources.Version) })

	ebiten.SetWindowTitle(pGame.Title())
	ebiten.SetWindowSize(
		pGame.Drawer().Game.Screen.Width,
		pGame.Drawer().Game.Screen.Height,
	)

	if err := ebiten.RunGame(pGame.Drawer()); err != nil {
		log.Fatal(err)
	}
}

type onlineMode struct {
	addr   string
	server bool
}

func (o onlineMode) gameMode() pkg.GameMode {
	return genericsutil.When[bool, pkg.GameMode](
		o.server, func(isServer bool) bool { return isServer },
		func(b bool) pkg.GameMode { return pkg.RemoteServerMode },
		func() pkg.GameMode { return pkg.RemoteClientMode })
}

func parseOnlineModeParam(server, client string) *onlineMode {
	if genericsutil.ForAll[string](func(s string) bool { return s == "" }, server, client) {
		return nil
	}

	return genericsutil.When[string, *onlineMode](
		server, func(addr string) bool { return addr != "" },
		func(s string) *onlineMode { return &onlineMode{server, true} },
		func() *onlineMode { return &onlineMode{client, false} })
}
