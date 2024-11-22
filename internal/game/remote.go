package game

import (
	"log"

	"github.com/joakim-ribier/pong/internal/drawer"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/internal/network/udp"
	"github.com/joakim-ribier/pong/pkg"
)

type RemotePongGame struct {
	GameDrawer *drawer.GameDrawer

	server   network.PServer
	client   network.PClient
	messages chan network.Message
	version  string
}

func NewRemotePongGame(
	debug bool,
	mode pkg.GameMode,
	remoteAddr, version string) *RemotePongGame {

	pg := &RemotePongGame{
		messages: make(chan network.Message),
		version:  version,
	}

	go pg.handleMessage()

	pg.GameDrawer = drawer.NewDrawerGame(
		pkg.NewGame(mode, debug),
		pg.Send, pg.shutdown,
		version)

	if pg.GameDrawer.Game.IsRemoteServer() {
		pg.server = udp.NewServer(remoteAddr)
		go pg.server.ListenAndServe(pg.messages)
		go pg.GameDrawer.Game.PlayerR.Remote()
	} else if pg.GameDrawer.Game.IsRemoteClient() {
		pg.client = udp.NewClient(remoteAddr)
		go pg.client.ListenAndServe(pg.messages)
		go pg.GameDrawer.Game.PlayerL.Remote()
	}

	return pg
}

// Drawer returns the drawer that builds the game
func (pg *RemotePongGame) Drawer() *drawer.GameDrawer {
	return pg.GameDrawer
}

// Title builds and returns the console title according to the remote type
func (pg *RemotePongGame) Title() string {
	title := pg.GameDrawer.Game.Title.Text
	if pg.GameDrawer.Game.IsRemoteServer() {
		title += " (Player L - server)"
	}
	if pg.GameDrawer.Game.IsRemoteClient() {
		title += " (Player R - client)"
	}
	return title
}

// handleMessage handles messages received from the web socket (server / client)
func (pg *RemotePongGame) handleMessage() {
	for message := range pg.messages {
		log.Printf("handle message from %s: %v", message.NetworkAddr, message)
		pg.GameDrawer.HandleNetworkMessage(message)
	}
}

func (pg *RemotePongGame) Send(msg network.Message) {
	if pg.GameDrawer.Game.IsRemoteServer() {
		pg.server.Send(msg)
	}
	if pg.GameDrawer.Game.IsRemoteClient() {
		pg.client.Send(msg)
	}
}

// shutdown turns off the application
func (pg *RemotePongGame) shutdown() bool {
	if pg.GameDrawer.Game.IsRemoteServer() {
		pg.server.Shutdown()
	}
	if pg.GameDrawer.Game.IsRemoteClient() {
		pg.client.Shutdown()
	}
	return true
}
