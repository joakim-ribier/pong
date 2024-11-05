package game

import (
	"github.com/joakim-ribier/pong/internal/drawer"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/internal/network/ws"
	"github.com/joakim-ribier/pong/pkg"
)

type RemotePongGame struct {
	GameDrawer *drawer.GameDrawer

	server   *ws.PServer
	client   *ws.PClient
	messages chan network.Message
}

func NewRemotePongGame(debug bool, mode pkg.GameMode, remoteAddr string) *RemotePongGame {
	pg := &RemotePongGame{
		messages: make(chan network.Message)}

	go pg.handleMessage()

	pg.GameDrawer = drawer.NewDrawerGame(
		pkg.NewGame(mode, debug),
		pg.Send,
		pg.shutdown)

	if pg.GameDrawer.Game.IsRemoteServer() {
		pg.server = ws.NewPServer(remoteAddr)
		go pg.server.ListenAndServe(pg.messages)
		go pg.GameDrawer.Game.PlayerR.Remote()
	} else if pg.GameDrawer.Game.IsRemoteClient() {
		pg.client = ws.NewPClient(remoteAddr)
		go pg.client.Conn(pg.messages)
		go pg.GameDrawer.Game.PlayerL.Remote()
		go pg.GameDrawer.Game.Ball.Remote()
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
		//log.Printf("received message from %s: %v", message.Id, message)
		pg.GameDrawer.NotifyRemoteMessage(message)
		switch message.Data.Cmd {
		case "connectionClosed":
			pg.GameDrawer.UpdateCurrentState(pkg.StartGame, true)
		case "currentState":
			if pg.GameDrawer.Game.IsRemoteClient() {
				pg.GameDrawer.UpdateCurrentState(pkg.ToState(message.Data.Value.(string)), true)
			}
		case "ping":
			if pg.GameDrawer.Game.IsRemoteServer() {
				pg.server.SendTo(message.Id, network.NewMessage("pong", nil))
			}
			if pg.GameDrawer.Game.IsRemoteClient() {
				pg.client.Send(network.NewMessage("pong", nil))
			}
		case "updateBallPosition":
			pg.GameDrawer.BallDrawer.UpdateBallPosition(message.Data.Value)
		case "updatePaddleY":
			pg.GameDrawer.PlayersDrawer.UpdatePaddleY(message.Data.Value)
		}
	}
}

func (pg *RemotePongGame) Send(cmd string, data interface{}) {
	switch cmd {
	case "updateBallPosition":
		if pg.GameDrawer.Game.IsRemoteServer() {
			pg.server.Send(network.NewMessage(cmd, data))
		}
	case "updatePaddleY":
		if pg.GameDrawer.Game.IsRemoteServer() {
			pg.server.Send(network.NewMessage(cmd, data))
		}
		if pg.GameDrawer.Game.IsRemoteClient() {
			pg.client.Send(network.NewMessage(cmd, data))
		}
	default:
		if pg.GameDrawer.Game.IsRemoteServer() {
			msg := network.NewMessage(cmd, data)
			pg.GameDrawer.NotifyRemoteMessage(msg)
			pg.server.Send(msg)
		}
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
