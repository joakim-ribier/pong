package drawer

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/mapsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/pkg"
)

type GameDrawer struct {
	Game *pkg.Game

	BallDrawer    BallDrawer
	PlayersDrawer PlayersDrawer

	shutdown func()
	send     func(msg network.Message)
	version  string

	keys []ebiten.Key

	remoteData *networkData
}

func NewDrawerGame(
	game *pkg.Game,
	send func(msg network.Message),
	shutdown func(),
	version string) *GameDrawer {

	return &GameDrawer{
		Game:          game,
		shutdown:      shutdown,
		send:          send,
		version:       version,
		BallDrawer:    *NewBallDrawer(*game),
		PlayersDrawer: *NewPlayerDrawer(*game),
		remoteData:    newNetworkData()}
}

func (g *GameDrawer) Draw(screen *ebiten.Image) {
	// draw the logo and the title
	g.drawBackgroundZone(screen)

	// draw the ping-pong table
	g.drawBackgroundGameZone(screen)

	// draw the debug info
	if g.Game.Debug {
		var b bytes.Buffer
		b.WriteString(fmt.Sprintf("Ticks per secondes: %0.2f", ebiten.ActualTPS()))
		b.WriteString(fmt.Sprintf("\nScreen W.H: %d %d", screen.Bounds().Size().X, screen.Bounds().Size().Y))
		b.WriteString("\nGame current state: " + g.Game.CurrentState.String())
		b.WriteString(fmt.Sprintf("\nBall X.Y: %0.2f %0.2f", g.Game.Ball.X, g.Game.Ball.Y))
		b.WriteString(fmt.Sprintf("\nPlayer L X.Y: %0.2f %0.2f", g.Game.PlayerL.Paddle.X, g.Game.PlayerL.Paddle.Y))
		b.WriteString(fmt.Sprintf("\nPlayer R X.Y: %0.2f %0.2f", g.Game.PlayerR.Paddle.X, g.Game.PlayerR.Paddle.Y))
		for nb, set := range g.Game.Win.Sets {
			if !set.EndTime.IsZero() {
				b.WriteString(fmt.Sprintf("\nSet n°%d: Time=%s, Speed=%0.02f, PlayerL=%d, PlayerR=%d, Win=%s",
					nb+1,
					set.Duration(),
					set.Speed(),
					set.PlayerLScore,
					set.PlayerRScore,
					set.PlayerSideWin.String()))
			}
		}
		ebitenutil.DebugPrint(screen, b.String())
	}

	// draw the "how to play" and player's score
	g.drawGameZoneText(screen)

	// draw the paddles
	if g.Game.CurrentState != pkg.WinGame {
		g.PlayersDrawer.Draw(screen)
	}

	// draw the ball
	if g.Game.CurrentState != pkg.ResumeGame && g.Game.CurrentState != pkg.WinGame {
		g.BallDrawer.Draw(screen)
	}

	// draw the counter zone between each set 3..2..1
	if g.Game.CurrentState == pkg.ResumeGame {
		tps := ebiten.ActualTPS()

		if int(float64(g.Game.ResumeGameState.Count)/tps) >= g.Game.ResumeGameState.Max {
			g.Game.CurrentState = pkg.PlayGame
		} else {
			remainingTime := g.Game.ResumeGameState.Max - int(float64(g.Game.ResumeGameState.Count)/tps)
			if remainingTime >= 0 {
				displayRemainingTime := fmt.Sprintf("%d", remainingTime)
				g.Game.ResumeGameState.Count++

				DrawText(screen, displayRemainingTime, g.Game.Screen.Font.H2, g.Game.Screen.AvailableColors["white"],
					pkg.Position{
						X: float32(int(g.Game.Screen.XRight)-(len(displayRemainingTime)*g.Game.Screen.Font.H2Size)) - 50,
						Y: float32(int(g.Game.Screen.YBottom)) + 25},
				)
			}
		}
	}

	// draw other info during a playing set...
	if g.Game.CurrentState == pkg.PlayGame || g.Game.CurrentState == pkg.ResumeGame || g.Game.CurrentState == pkg.PauseGame {
		posX := float32(g.Game.Screen.XLeft + 50)
		posY := float32(g.Game.Screen.YBottom + 30)
		marginY := float32(g.Game.Screen.Font.SmallTextSize + 10)
		marginX := float32(60)
		font := g.Game.Screen.Font.SmallText
		fontSize := g.Game.Screen.Font.SmallTextSize
		color := g.Game.Screen.AvailableColors["white"]

		currentSet := g.Game.CurrentSet()
		lastSet := g.Game.LastEndedSet()

		time := time.Unix(0, 0).UTC().Add(time.Since(currentSet.StartTime).Round(time.Second)).Format("04:05")
		DrawText(screen, "Time:", font, color, pkg.Position{X: posX, Y: posY})
		DrawText(screen, time, font, color, pkg.Position{X: posX + marginX, Y: posY})
		posY += marginY

		ballSpeed := fmt.Sprintf("%0.02f", currentSet.XSpeed)
		DrawText(screen, "Speed:", font, color, pkg.Position{X: posX, Y: posY})
		DrawText(screen, ballSpeed, font, color, pkg.Position{X: posX + marginX, Y: posY})
		if lastSet != nil {
			DrawText(screen, fmt.Sprintf("(%0.02f)", lastSet.XSpeed), font, color, pkg.Position{X: posX + marginX + float32(GetSize(ballSpeed, fontSize)) + 5, Y: posY})
		}
		posY += marginY

		nbHit := fmt.Sprintf("%d", currentSet.NbHit)
		DrawText(screen, "Hits:", font, color, pkg.Position{X: posX, Y: posY})
		DrawText(screen, nbHit, font, color, pkg.Position{X: posX + marginX, Y: posY})
		if lastSet != nil {
			DrawText(screen, fmt.Sprintf("(%d)", lastSet.NbHit), font, color, pkg.Position{X: posX + marginX + float32(GetSize(nbHit, fontSize)) + 5, Y: posY})
		}

		posY += marginY
	}

	// draw the winner zone
	if g.Game.CurrentState == pkg.WinGame {
		g.drawWinnerGameZone(screen)
	}

	// draw the remote info game zone
	if !g.Game.IsLocal() {
		g.drawRemoteGameZone(screen)
	}
}

func (g *GameDrawer) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	// handle the shutdown by the user CTRL+C
	shutdown := len(slicesutil.FilterT[ebiten.Key](g.keys, func(k ebiten.Key) bool {
		return k == ebiten.KeyControl || k == ebiten.KeyC
	})) == 2
	if shutdown {
		g.shutdown()
		return fmt.Errorf("shutdown app '%s' now", g.Game.Title.Text)
	}

	// draw the ping-pong table on the start screen as a logo
	if g.Game.CurrentState == pkg.StartGame {
		screen := g.Game.Screen
		screen.YBottom = float32(g.Game.Screen.GameZoneYCenter()) - float32(g.Game.PlayerL.Paddle.Height/2) - 15
		screen.YTop = float32(g.Game.Screen.GameZoneYCenter()) + float32(g.Game.PlayerL.Paddle.Height/2) + 15
		screen.XLeft = float32(g.Game.Screen.GameZoneXCenter()) - 150
		screen.XRight = float32(g.Game.Screen.GameZoneXCenter()) + 150

		g.PlayersDrawer.UpdateLeft(screen)
		g.PlayersDrawer.UpdateRight(screen)
		g.BallDrawer.Update(g.Game, screen, true)
	}

	if g.Game.CurrentState == pkg.PlayGame {
		g.BallDrawer.Update(g.Game, g.Game.Screen, false)
	}

	if g.Game.CurrentState == pkg.PlayGame || g.Game.CurrentState == pkg.ResumeGame {
		g.updatePlayer(g.BallDrawer.playerL)
		g.updatePlayer(g.BallDrawer.playerR)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !g.Game.IsRemoteClient() {
		switch g.Game.CurrentState {
		case pkg.PlayGame:
			g.updateCurrentState(pkg.PauseGame)
		case pkg.StartGame:
			if g.Game.IsRemoteServer() && !g.remoteData.readyToPlay.ready {
				g.addMessageWithLevel(fmt.Sprintf("%s is not ready", g.Game.PlayerR.Name), warning)
			} else {
				g.updateCurrentState(pkg.ResumeGame)
			}
		case pkg.WinGame:
			g.updateCurrentState(pkg.StartGame)
		default:
			g.updateCurrentState(pkg.PlayGame)
		}
	}

	if g.Game.IsRemoteClient() && g.Game.CurrentState == pkg.StartGame {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.remoteData.readyToPlay.ready = !g.remoteData.readyToPlay.ready
			g.remoteData.readyToPlay.nbTpsLaps = 0

			// send the cmd to the remote side
			g.send(network.NewMessage(network.Ready.String(), g.remoteData.readyToPlay.ready))
		}
		if g.remoteData.readyToPlay.ready {
			nbSeconds := computeSecondFromTPS(g.remoteData.readyToPlay.nbTpsLaps)
			g.remoteData.readyToPlay.nbSeconds = nbSeconds
			if nbSeconds > g.remoteData.readyToPlay.nbSecondsMax {
				g.remoteData.readyToPlay.nbTpsLaps = -1
			}
			g.remoteData.readyToPlay.nbTpsLaps += 1
		}
	}

	return nil
}

func (g *GameDrawer) updatePlayer(player pkg.Player) {
	currentY := player.Paddle.Y
	if state := g.PlayersDrawer.update(player, g.Game.Screen); state.PlayerLostBall() {
		g.updateCurrentState(state)
	}
	if currentY != player.Paddle.Y {
		g.send(network.NewMessage(network.UpdatePaddleY.String(), player.Paddle.Y))
	}
}

func (g *GameDrawer) updateCurrentState(state pkg.State) {
	g.Game.CurrentState = state
	switch state {
	case pkg.PlayerLLostBall:
		if g.Game.IsRemoteServer() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
		g.addMessageWithLevel("Player R wins the point", logg)
		g.playerWinSet(g.Game.PlayerR)
	case pkg.PlayerRLostBall:
		if g.Game.IsRemoteClient() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
		g.addMessageWithLevel("Player L wins the point", logg)
		g.playerWinSet(g.Game.PlayerL)
	case pkg.PauseGame:
		if g.Game.IsRemoteServer() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
		g.addMessageWithLevel("Pause game...", logg)
	case pkg.PlayGame:
		if g.Game.IsRemoteServer() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
	case pkg.ResumeGame:
		if g.Game.IsLocal() {
			g.Game.StartNewSet()
			return
		}
		if g.Game.IsRemoteServer() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
		if size := len(g.Game.Win.Sets); size == 0 {
			g.addMessageWithLevel("Start a new game", logg)
		}
		g.Game.StartNewSet()
		g.addMessageWithLevel(fmt.Sprintf("Start new set (%d)", len(g.Game.Win.Sets)), logg)
	case pkg.StartGame:
		g.remoteData.readyToPlay.ready = false
		if g.Game.IsRemoteServer() {
			g.send(network.NewMessage(network.UpdateCurrentState.String(), g.Game.CurrentState.String()))
		}
		g.Game.ResetGame()
	case pkg.WinGame:
		g.remoteData.readyToPlay.ready = false
		g.addMessageWithLevel("End of the game", logg)
		if player := g.Game.Winner(); player != nil {
			g.addMessageWithLevel(fmt.Sprintf("%s wins! (%d/%d)", player.Name, player.Score, g.Game.Looser().Score), info)
		}
	}
}

func (g *GameDrawer) HandleNetworkMessage(message network.Message) {
	switch message.AsCMD() {
	case network.Notify:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.addMessageWithLevel(fmt.Sprintf("%s", message.Data.Value), info)
		}
	case network.Ping:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.send(network.NewMessage(network.Pong.String(), g.version).WithAddr(message.NetworkAddr))
		}
	case network.PingAll:
		for _, client := range g.remoteData.clients {
			if client.nbPingAttempts >= client.nbPingMaxAttempts {
				// delete the subscriber if it not responding...
				g.HandleNetworkMessage(network.NewSimpleMessage(network.Shutdown.String()).WithAddr(client.networkAddr))
				return
			}
			client.lastPing = time.Now()
			client.nbPingAttempts += 1
		}
	case network.Pong:
		if client, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			client.lastPong = time.Now()
			client.nbPingAttempts = 0
			client.version = message.Data.Value.(string)
		}
	case network.Ready:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.remoteData.readyToPlay.ready = message.Data.Value.(bool)
			if g.Game.IsRemoteServer() {
				if g.remoteData.readyToPlay.ready {
					g.addMessageWithLevel(fmt.Sprintf("%s ready to play", message.NetworkAddr), info)
				} else {
					g.addMessageWithLevel(fmt.Sprintf("%s not ready anymore", message.NetworkAddr), warning)
				}
			}
		}
	case network.Shutdown:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.addMessageWithLevel("Lost connection...", warning)
			g.addMessageWithLevel(fmt.Sprintf("%s disconnected", message.NetworkAddr), warning)
			delete(g.remoteData.clients, message.NetworkAddr)
			g.updateCurrentState(pkg.StartGame)
		}
	case network.Subscribe:
		if g.Game.IsRemoteClient() {
			if message.Data.Value.(string) != "accepted" {
				g.addMessageWithLevel("Connection refused...", warning)
				return
			}
			g.addMessageWithLevel("Press [space] to start...", info)
		} else {
			if len(g.remoteData.clients) > 0 {
				g.send(network.NewMessage(network.Subscribe.String(), "refused").WithAddr(message.NetworkAddr))
				return
			}
			g.addMessageWithLevel("New subscriber...", logg)
			g.addMessageWithLevel(fmt.Sprintf("%s connected", message.NetworkAddr), logg)
			g.send(network.NewMessage(network.Subscribe.String(), "accepted").WithAddr(message.NetworkAddr))
		}
		g.remoteData.clients[message.NetworkAddr] = newRemoteClient(message.NetworkAddr)
		g.remoteData.clients[message.NetworkAddr].lastPing = time.Now()
	case network.UpdateCurrentState:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.updateCurrentState(pkg.ToState(message.Data.Value.(string)))
		}
	case network.UpdatePaddleY:
		if _, ok := g.remoteData.clients[message.NetworkAddr]; ok {
			g.PlayersDrawer.UpdatePaddleY(message.Data.Value)
		}
	}
}

func (g *GameDrawer) playerWinSet(player *pkg.Player) {
	g.Game.Mark(player)
	g.Game.EndSet(*player)
	if g.Game.Winner() != nil {
		g.updateCurrentState(pkg.WinGame)
	} else {
		if !g.Game.IsRemoteClient() {
			g.updateCurrentState(pkg.ResumeGame)
		}
	}
}

// drawBackgroundZone draws the background (logo + title)
func (g *GameDrawer) drawBackgroundZone(screen *ebiten.Image) {
	titleSize := len(g.Game.Title.Text) * g.Game.Title.FontSize
	titleX := float32(GetXCenterPos(g.Game.Screen.Width, g.Game.Title.Text, int(g.Game.Title.FontSize)))
	titleXMargin := 20

	totalLogoSize := 110
	yBottom := int(g.Game.Screen.YBottom) - (int(g.Game.Screen.YBottom) - totalLogoSize)
	xLeft := int(g.Game.Screen.XLeft) - (int(g.Game.Screen.XLeft) - totalLogoSize)

	/*
		|| ---------      --------
		|| || ------ PONG --------
		|| || || ---      --------
		|| || ||
		|| || ||
	*/
	drawAtariLogo := func() {
		// left lines
		DrawRectangle(screen,
			int(titleX)-titleXMargin-10, 20,
			pkg.Position{X: float32(xLeft/2) - 40, Y: float32(yBottom/2) - 40},
			g.Game.Screen.AvailableColors["#atari1"])

		DrawRectangle(screen,
			int(titleX)-titleXMargin-40, 20,
			pkg.Position{X: float32(xLeft/2) - 10, Y: float32(yBottom/2) - 10},
			g.Game.Screen.AvailableColors["#atari2"])

		DrawRectangle(screen,
			int(titleX)-titleXMargin-70, 20,
			pkg.Position{X: float32(xLeft/2) + 20, Y: float32(yBottom/2) + 20},
			g.Game.Screen.AvailableColors["#atari3"])

		// right lines
		DrawRectangle(screen,
			g.Game.Screen.Width/2, 20,
			pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(yBottom/2) - 40},
			g.Game.Screen.AvailableColors["#atari1"])

		DrawRectangle(screen,
			g.Game.Screen.Width/2, 20,
			pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(yBottom/2) - 10},
			g.Game.Screen.AvailableColors["#atari2"])

		DrawRectangle(screen,
			g.Game.Screen.Width/2, 20,
			pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(yBottom/2) + 20},
			g.Game.Screen.AvailableColors["#atari3"])

		// vertical lines
		DrawRectangle(screen,
			20, g.Game.Screen.Height,
			pkg.Position{X: float32(xLeft/2) - 40, Y: float32(yBottom/2) - 20},
			g.Game.Screen.AvailableColors["#atari1"])

		DrawRectangle(screen,
			20, g.Game.Screen.Height,
			pkg.Position{X: float32(xLeft/2) - 10, Y: float32(yBottom / 2)},
			g.Game.Screen.AvailableColors["#atari2"])

		DrawRectangle(screen,
			20, g.Game.Screen.Height,
			pkg.Position{X: float32(xLeft/2) + 20, Y: float32(yBottom/2) + 20},
			g.Game.Screen.AvailableColors["#atari3"])
	}

	drawTitle := func() {
		DrawText(screen, "#"+g.version, g.Game.Screen.Font.TinyText, g.Game.Subtitle.Color,
			pkg.Position{
				X: float32(GetXCenterPos(g.Game.Screen.Width, "#"+g.version, g.Game.Screen.Font.TinyTextSize)),
				Y: float32(yBottom/2) - float32(g.Game.Title.FontSize)/2 - float32(g.Game.Screen.Font.TinyTextSize) - 15},
		)

		DrawText(screen, g.Game.Title.Text, g.Game.Title.Font, g.Game.Title.Color,
			pkg.Position{
				X: float32(GetXCenterPos(g.Game.Screen.Width, g.Game.Title.Text, g.Game.Title.FontSize)),
				Y: float32(yBottom/2) - float32(g.Game.Title.FontSize)/2},
		)

		DrawText(screen, g.Game.Subtitle.Text, g.Game.Subtitle.Font, g.Game.Subtitle.Color,
			pkg.Position{
				X: float32(GetXCenterPos(g.Game.Screen.Width, g.Game.Subtitle.Text, g.Game.Subtitle.FontSize)),
				Y: float32(yBottom/2) + float32(g.Game.Title.FontSize)/2 - float32(g.Game.Subtitle.FontSize/2) + 15},
		)
	}

	screen.Fill(g.Game.Screen.AvailableColors["#bg"])
	drawTitle()
	drawAtariLogo()
}

// drawGameZoneTextZone draws the remote text info
func (g *GameDrawer) drawRemoteGameZone(screen *ebiten.Image) {
	textColor := g.Game.Screen.AvailableColors["white"]
	font := g.Game.Screen.Font.TinyText
	fonSize := g.Game.Screen.Font.TinyTextSize
	marginY := 2

	DrawRectangle(screen, g.Game.Screen.RemoteExtendZoneW-15, g.Game.Screen.GameZoneHeight()+30, pkg.Position{
		X: float32(g.Game.Screen.XLeft - float32(g.Game.Screen.RemoteExtendZoneW)),
		Y: g.Game.Screen.YBottom - 15}, color.Black)

	y := g.Game.Screen.YBottom

	text := genericsutil.When[bool, string](
		g.Game.IsRemoteServer(), func(b bool) bool { return b },
		func(b bool) string { return "[CLIENT ADDR] / [PING]" },
		func() string { return "[SERVER ADDR] / [PING]" })

	DrawText(screen, text, font, textColor,
		pkg.Position{
			X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW/2)) - float32(len(text)*fonSize)/2,
			Y: y},
	)
	y += float32(g.Game.Screen.Font.TextSize) + float32(marginY)

	for _, client := range mapsutil.Sort(g.remoteData.clients) {
		text = client.networkAddr
		DrawText(screen, text, font, textColor,
			pkg.Position{
				X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW/2)) - float32(len(text)*fonSize)/2,
				Y: y},
		)
		y += float32(g.Game.Screen.Font.TextSize) + float32(marginY)

		text = "#" + client.version
		DrawText(screen, text, font, textColor,
			pkg.Position{
				X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW/2)) - float32(GetSize(text, fonSize))/2,
				Y: y},
		)
		y += float32(g.Game.Screen.Font.TextSize) + float32(marginY)

		text = "..."
		if !client.lastPong.IsZero() {
			nbAttemptsText := ""
			if client.nbPingAttempts > 1 {
				nbAttemptsText = fmt.Sprintf(" (%d/%d)", client.nbPingAttempts, client.nbPingMaxAttempts)
			}
			text = fmt.Sprintf("%s%s - %d ms",
				client.lastPong.Format("15:04:05"),
				nbAttemptsText,
				client.ping().Milliseconds())
		}
		DrawText(screen, text, font, textColor,
			pkg.Position{
				X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW/2)) - float32(len(text)*fonSize)/2,
				Y: y},
		)
		y += float32(g.Game.Screen.Font.TextSize) + float32(marginY*3)
	}

	text = "[CHANNEL]"
	DrawText(screen, text, font, textColor,
		pkg.Position{
			X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW)) + 5,
			Y: y},
	)
	y += float32(g.Game.Screen.Font.TextSize) + float32(marginY)

	for _, msg := range g.remoteData.messages {
		DrawText(screen, msg.dateTime, font, textColor,
			pkg.Position{
				X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW)) + 5,
				Y: y},
		)

		DrawText(screen, msg.text, font, msg.asColor(g.Game.Screen.AvailableColors),
			pkg.Position{
				X: float32(g.Game.Screen.XLeft-float32(g.Game.Screen.RemoteExtendZoneW)) + 80,
				Y: y},
		)
		y += float32(g.Game.Screen.Font.TextSize) + float32(marginY)
	}

	if g.Game.IsRemoteClient() {
		if g.remoteData.readyToPlay.ready && g.Game.CurrentState == pkg.StartGame {
			font := g.Game.Screen.Font.H2
			fontSize := g.Game.Screen.Font.H2Size

			DrawRectangle(screen,
				g.Game.Screen.GameZoneWidth(), g.Game.Screen.GameZoneHeight(),
				pkg.Position{X: float32(g.Game.Screen.XLeft), Y: float32(g.Game.Screen.YBottom)},
				color.RGBA{0, 0, 0, 200})

			textZoneW := 400
			textZoneH := 100
			DrawRectangle(screen,
				textZoneW, textZoneH,
				pkg.Position{
					X: float32(g.Game.Screen.GameZoneXCenter() - textZoneW/2),
					Y: float32(g.Game.Screen.GameZoneYCenter() - textZoneH/2)},
				color.Black)

			text := "READY"
			if nbSeconds := g.remoteData.readyToPlay.nbSeconds; nbSeconds > 0 && nbSeconds <= g.remoteData.readyToPlay.nbSecondsMax {
				text = fmt.Sprintf("%s%s", text, strings.Repeat(".", nbSeconds))
			}

			DrawText(screen, text, font, color.White,
				pkg.Position{
					X: float32(g.Game.Screen.GameZoneXCenter()) - float32(GetSize(text, fontSize))/2,
					Y: float32(g.Game.Screen.GameZoneYCenter() - fontSize/2)},
			)

			font = g.Game.Screen.Font.SmallText
			fontSize = g.Game.Screen.Font.SmallTextSize
			text = "Please wait for the server to start the game."
			DrawText(screen, text, font, color.White,
				pkg.Position{
					X: float32(g.Game.Screen.GameZoneXCenter()) - float32(GetSize(text, fontSize))/2,
					Y: float32(g.Game.Screen.GameZoneYCenter() + fontSize/2 + 20)},
			)
		}
	}
}

func computeSecondFromTPS(nb int) int {
	tps := ebiten.ActualTPS()
	if tps > 0 {
		millis := nb * (1000 / int(tps))
		s := 0
		if millis >= 1000 && millis < 2000 {
			s = 1
		} else if millis >= 2000 && millis < 3000 {
			s = 2
		} else if millis >= 3000 && millis < 4000 {
			s = 3
		} else if millis >= 4000 {
			s = 4
		}
		return s
	} else {
		return 0
	}
}

// drawGameZoneText draws the "how to play" and the players's score zone (name and score)
func (g *GameDrawer) drawGameZoneText(screen *ebiten.Image) {
	marginCenterX := 45
	marginTopY := float32(35)

	playerLText := fmt.Sprintf("%s: %d", strings.ToUpper(g.Game.PlayerL.Name), g.Game.PlayerL.Score)

	// Player L score
	DrawText(screen, playerLText, g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
		pkg.Position{
			X: float32(g.Game.Screen.GameZoneXCenter() - (len(playerLText) * g.Game.Screen.Font.TextSize) - marginCenterX),
			Y: g.Game.Screen.YBottom + marginTopY},
	)

	// Player R score
	DrawText(screen, fmt.Sprintf("%s: %d", strings.ToUpper(g.Game.PlayerR.Name), g.Game.PlayerR.Score), g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
		pkg.Position{
			X: float32(g.Game.Screen.GameZoneXCenter() + marginCenterX),
			Y: g.Game.Screen.YBottom + marginTopY},
	)

	// display the user's info - how to play it
	if g.Game.CurrentState == pkg.StartGame {
		y := int(g.Game.Screen.YBottom) + int(marginTopY) + 40

		description := []string{}
		description = append(description,
			"# THE WINNER",
			"",
			fmt.Sprintf("The first player to %d points", g.Game.Win.SetScore),
			fmt.Sprintf("with %d points difference", g.Game.Win.SetGapWScore),
			"",
			fmt.Sprintf("or the first to %d points!", g.Game.Win.Score))

		for _, line := range description {
			DrawText(screen, line, g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
				pkg.Position{
					X: float32(g.Game.Screen.XLeft) + 35,
					Y: float32(y)},
			)
			y += g.Game.Screen.Font.TextSize + 10
		}

		y = int(g.Game.Screen.YBottom) + int(marginTopY) + 40
		description = []string{}
		description = append(description,
			"# HOW TO PLAY",
			"",
			"Player L -> Z + S",
			"Player R -> UP + DOWN",
			"")
		if g.Game.IsRemoteClient() {
			description = append(description, "Press [space] when you are ready", "to start the game...")
		} else {
			description = append(description, "Press [space] to start or pause", "at every moment...")
		}
		for _, line := range description {
			DrawText(screen, line, g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
				pkg.Position{
					X: float32(g.Game.Screen.GameZoneXCenter()) + 20,
					Y: float32(y)},
			)
			y += g.Game.Screen.Font.TextSize + 10
		}
	}
}

// drawBackgroundGameZone draws the specific game zone inside the screen
func (g *GameDrawer) drawBackgroundGameZone(screen *ebiten.Image) {
	DrawRectangle(screen,
		g.Game.Screen.GameZoneWidth(), g.Game.Screen.GameZoneHeight(),
		pkg.Position{X: float32(g.Game.Screen.XLeft), Y: float32(g.Game.Screen.YBottom)},
		g.Game.Screen.AvailableColors["#table-bg"])

	DrawRectangle(screen,
		g.Game.Screen.GameZoneWidth()-10, g.Game.Screen.GameZoneHeight()-10,
		pkg.Position{X: float32(g.Game.Screen.XLeft) + 5, Y: float32(g.Game.Screen.YBottom + 5)},
		g.Game.Screen.AvailableColors["white"])

	DrawRectangle(screen,
		g.Game.Screen.GameZoneWidth()-40, g.Game.Screen.GameZoneHeight()-30,
		pkg.Position{X: float32(g.Game.Screen.XLeft) + 20, Y: float32(g.Game.Screen.YBottom + 15)},
		g.Game.Screen.AvailableColors["#table-bg"])

	vector.DrawFilledCircle(screen, float32(g.Game.Screen.GameZoneXCenter()), float32(g.Game.Screen.GameZoneYCenter()), float32(70), g.Game.Screen.AvailableColors["white"], true)
	vector.DrawFilledCircle(screen, float32(g.Game.Screen.GameZoneXCenter()), float32(g.Game.Screen.GameZoneYCenter()), float32(65), g.Game.Screen.AvailableColors["#bg"], true)
	vector.DrawFilledCircle(screen, float32(g.Game.Screen.GameZoneXCenter()), float32(g.Game.Screen.GameZoneYCenter()), float32(30), g.Game.Screen.AvailableColors["white"], true)
	vector.DrawFilledCircle(screen, float32(g.Game.Screen.GameZoneXCenter()), float32(g.Game.Screen.GameZoneYCenter()), float32(25), g.Game.Screen.AvailableColors["#bg"], true)

	DrawRectangle(screen,
		5, g.Game.Screen.GameZoneHeight()-20,
		pkg.Position{X: float32(g.Game.Screen.GameZoneXCenter()) - 2.5, Y: g.Game.Screen.YBottom + 10},
		g.Game.Screen.AvailableColors["white"])
}

// drawWinnerGameZone draws the winner player zone and game stats
func (g *GameDrawer) drawWinnerGameZone(screen *ebiten.Image) {
	if player := g.Game.Winner(); player != nil {
		textSize := g.Game.Screen.Font.H2Size
		textFont := g.Game.Screen.Font.H2

		marginTextSize := 10
		marginTitleSize := 150

		x := int(g.Game.Screen.GameZoneXCenter())
		direction := g.Game.Screen.GameZoneWidth() / 4
		if player.Side == pkg.PlayerLeft {
			direction = direction * -1
		}

		// draw the Winner info
		DrawText(screen, "And the winner", textFont, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(x+direction) - float32((len("And the winner")*textSize))/2,
				Y: float32(g.Game.Screen.GameZoneYCenter() - marginTitleSize)},
		)
		DrawText(screen, "is...", textFont, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(x+direction) - float32((len("is...")*textSize))/2,
				Y: float32(g.Game.Screen.GameZoneYCenter() - marginTitleSize + marginTextSize + textSize)},
		)

		DrawText(screen, strings.ToUpper(player.Name), textFont, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(x+direction) - float32((len(player.Name)*textSize))/2,
				Y: float32(g.Game.Screen.GameZoneYCenter() - textSize/2)},
		)

		// draw the stats players
		DrawText(screen, "Stats", textFont, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(x+(direction*-1)) - float32((len("Stats")*textSize))/2,
				Y: float32(g.Game.Screen.GameZoneYCenter() - marginTitleSize)},
		)

		if len(g.Game.Win.Sets) > 0 {
			totalTime := g.Game.Win.Sets[len(g.Game.Win.Sets)-1].EndTime.Sub(g.Game.Win.Sets[0].StartTime)
			totalTimeToDisplay := fmt.Sprintf("Time: %s", time.Unix(0, 0).UTC().Add(totalTime).Format("04:05"))
			DrawText(screen, totalTimeToDisplay, g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
				pkg.Position{
					X: float32(x+(direction*-1)) - float32((len(totalTimeToDisplay)*g.Game.Screen.Font.TextSize))/2,
					Y: float32(g.Game.Screen.GameZoneYCenter() - g.Game.Screen.Font.TextSize - marginTextSize)},
			)

			maxXSpeedToDisplay := fmt.Sprintf("Max speed: %0.02f", g.Game.MaxXSpeedSet())
			DrawText(screen, maxXSpeedToDisplay, g.Game.Screen.Font.Text, g.Game.Screen.AvailableColors["white"],
				pkg.Position{
					X: float32(x+(direction*-1)) - float32((len(maxXSpeedToDisplay)*g.Game.Screen.Font.TextSize))/2,
					Y: float32(g.Game.Screen.GameZoneYCenter() + marginTextSize)},
			)
		}

		y := g.Game.Screen.GameZoneYCenter() + marginTextSize*3
		toTextSize := -1
		for _, set := range g.Game.Win.Sets {
			y += 15

			color := g.Game.Screen.AvailableColors["#bg"]
			if set.HasWin(*player) {
				color = g.Game.Screen.AvailableColors["white"]
			}

			textSize := g.Game.Screen.Font.SmallTextSize
			toText := fmt.Sprintf("T. %s X. %s", set.Duration(), set.SpeedFormat())
			if toTextSize == -1 {
				toTextSize = len(toText) * textSize
			}

			DrawRectangle(screen, textSize, textSize, pkg.Position{
				X: float32(x + direction - textSize/2 - textSize - toTextSize/2),
				Y: float32(y)}, color)

			DrawText(screen, toText, g.Game.Screen.Font.SmallText, g.Game.Screen.AvailableColors["white"],
				pkg.Position{
					X: float32(x + direction + textSize - toTextSize/2),
					Y: float32(y)},
			)
		}

	}
}

func (g *GameDrawer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Game.Screen.Width, g.Game.Screen.Height
}

func (g *GameDrawer) addMessageWithLevel(msg string, level networkMessageLevel) {
	maxSize := 40
	if len(g.remoteData.messages) > maxSize {
		g.remoteData.messages = g.remoteData.messages[len(g.remoteData.messages)-maxSize:]
	}

	g.remoteData.messages = append(
		g.remoteData.messages,
		networkMessage{dateTime: fmt.Sprintf("[%s]", time.Now().Format("15:04:05")), text: msg, level: level},
	)
}
