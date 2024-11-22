package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/pong/internal/game"
	"github.com/joakim-ribier/pong/pkg"
)

func main() {
	debug := flag.Bool("debug", false, "enable the 2-D engine [debug] mode")
	server := flag.String("server", "", "start a server [--server 0.0.0:3000] to host the game")
	client := flag.String("client", "", "start a client [--client 0.0.0:3000] to connect to the server")
	verbose := flag.Bool("verbose", false, "enable the [verbose] mode to display logs")
	dev := flag.Bool("dev", false, "start the application in [dev] mode")

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if !*verbose {
		log.SetOutput(io.Discard)
	}

	var pongGame game.PongGame
	if param := parseOnlineModeParam(*server, *client); param != nil {
		pongGame = game.NewRemotePongGame(*debug, param.AsGameMode(), param.addr, version(*dev))
	} else {
		pongGame = game.NewLocalPongGame(*debug, version(*dev))
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

type onlineMode struct {
	addr   string
	server bool
}

func (o onlineMode) AsGameMode() pkg.GameMode {
	return genericsutil.When[bool, pkg.GameMode](o.server,
		func(isServer bool) bool { return isServer },
		pkg.RemoteServerMode,
		pkg.RemoteClientMode)
}

func parseOnlineModeParam(server, client string) *onlineMode {
	if server == "" && client == "" {
		return nil
	}

	return genericsutil.When[string, *onlineMode](server,
		func(addr string) bool { return addr != "" },
		&onlineMode{server, true},
		&onlineMode{client, false})
}

func version(dev bool) string {
	if dev {
		cmd := exec.Command("git", "log", `--format=%h`, "-n1")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			_ = os.WriteFile("release.txt", out, 0644)
		}
	}

	version, err := os.ReadFile("release.txt")
	if err != nil {
		return "unknowm"
	}
	return string(version)
}
