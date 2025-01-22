package online

import (
	"fmt"
	"log"

	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/pong/internal/drawer"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/internal/network/udp"
	"github.com/joakim-ribier/pong/pkg"
)

type OnlinePGame struct {
	GameDrawer *drawer.GameDrawer

	server   network.Conn
	client   network.Conn
	messages chan network.Message
	version  string
}

func NewPGame(debug bool, mode pkg.GameMode, networkAddr, version string) *OnlinePGame {
	pg := &OnlinePGame{
		messages: make(chan network.Message),
		version:  version,
	}

	go pg.handleMessage()
	pg.GameDrawer = drawer.NewDrawerGame(pkg.NewGame(mode, debug), pg.send, pg.shutdown, version)

	if pg.GameDrawer.Game.IsRemoteServer() {
		pg.server = udp.NewServer(networkAddr)
		go pg.server.ListenAndServe(pg.messages)
		go pg.GameDrawer.Game.PlayerR.Remote()
	} else if pg.GameDrawer.Game.IsRemoteClient() {
		pg.client = udp.NewClient(networkAddr)
		go pg.client.ListenAndServe(pg.messages)
		go pg.GameDrawer.Game.PlayerL.Remote()
	}

	return pg
}

// Drawer returns the drawer that builds the game
func (pg *OnlinePGame) Drawer() *drawer.GameDrawer {
	return pg.GameDrawer
}

// Title builds and returns the console title according to the remote type
func (pg *OnlinePGame) Title() string {
	title := pg.GameDrawer.Game.Title.Text

	return genericsutil.When[pkg.Game, string](
		*pg.GameDrawer.Game, func(game pkg.Game) bool { return game.IsRemoteServer() },
		func(g pkg.Game) string {
			return fmt.Sprintf("%s (%s - server)", title, pg.GameDrawer.Game.PlayerL.Name)
		},
		func() string { return fmt.Sprintf("%s (%s - client)", title, pg.GameDrawer.Game.PlayerR.Name) })
}

// handleMessage handles messages received from the network
func (pg *OnlinePGame) handleMessage() {
	for message := range pg.messages {
		log.Printf("handle message from %s: %v", message.NetworkAddr, message)
		pg.GameDrawer.HandleNetworkMessage(message)
	}
}

// send sends a message to the network
func (pg *OnlinePGame) send(msg network.Message) {
	genericsutil.When[pkg.Game, network.Conn](
		*pg.GameDrawer.Game, func(game pkg.Game) bool { return game.IsRemoteServer() },
		func(g pkg.Game) network.Conn { return pg.server }, func() network.Conn { return pg.client }).Send(msg)
}

// shutdown turns off the network listener
func (pg *OnlinePGame) shutdown() {
	genericsutil.When[pkg.Game, network.Conn](
		*pg.GameDrawer.Game, func(game pkg.Game) bool { return game.IsRemoteServer() },
		func(g pkg.Game) network.Conn { return pg.server }, func() network.Conn { return pg.client }).Shutdown()
}
