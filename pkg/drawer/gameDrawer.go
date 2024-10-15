package drawer

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/pong/pkg"
)

type GameDrawer struct {
	Game *pkg.Game

	BallDrawer    BallDrawer
	PlayersDrawer PlayersDrawer
}

func NewDrawerGame(game *pkg.Game) *GameDrawer {
	return &GameDrawer{
		Game:          game,
		BallDrawer:    *NewBallDrawer(*game),
		PlayersDrawer: *NewPlayerDrawer(*game)}
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
	g.drawGameZoneTextZone(screen)

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
						X: float32(int(g.Game.Screen.XRight)-(len(displayRemainingTime)*g.Game.Screen.Font.H2Size)) - 30,
						Y: float32(int(g.Game.Screen.YBottom)) + 25},
				)
			}
		}
	}

	// draw other info during a playing set...
	if g.Game.CurrentState == pkg.PlayGame || g.Game.CurrentState == pkg.ResumeGame || g.Game.CurrentState == pkg.PauseGame {
		currentSet := g.Game.Win.Sets[len(g.Game.Win.Sets)-1]
		timeToDisplay := fmt.Sprintf("Time: %s", time.Unix(0, 0).UTC().Add(time.Since(currentSet.StartTime).Round(time.Second)).Format("04:05"))
		DrawText(screen, timeToDisplay, g.Game.Screen.Font.SmallText, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(g.Game.Screen.XLeft + 30),
				Y: float32(g.Game.Screen.YBottom + 30)},
		)

		ballSpeed := fmt.Sprintf("Speed: %0.02f", genericsutil.OrElse[float32](g.Game.Ball.XSpeed, func() bool { return g.Game.Ball.XSpeed > 0 }, g.Game.Ball.XSpeed*-1))
		DrawText(screen, ballSpeed, g.Game.Screen.Font.SmallText, g.Game.Screen.AvailableColors["white"],
			pkg.Position{
				X: float32(g.Game.Screen.XLeft + 30),
				Y: float32(int(g.Game.Screen.YBottom) + 30 + g.Game.Screen.Font.SmallTextSize + 10)},
		)
	}

	// draw the winner zone
	if g.Game.CurrentState == pkg.WinGame {
		g.drawWinnerGameZone(screen)
	}
}

func (g *GameDrawer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Game.Screen.Width, g.Game.Screen.Height
}

func (g *GameDrawer) Update() error {
	if g.Game.CurrentState == pkg.StartGame || g.Game.CurrentState == pkg.WinGame {
		// draw the ping-pong on the start screen as a logo
		screen := g.Game.Screen
		screen.YBottom = float32(g.Game.Screen.GameZoneYCenter()) - float32(g.Game.PlayerL.Paddle.Height/2) - 15
		screen.YTop = float32(g.Game.Screen.GameZoneYCenter()) + float32(g.Game.PlayerL.Paddle.Height/2) + 15
		screen.XLeft = float32(g.Game.Screen.GameZoneXCenter()) - 150
		screen.XRight = float32(g.Game.Screen.GameZoneXCenter()) + 150

		g.PlayersDrawer.Update(screen)

		if g.Game.CurrentState == pkg.StartGame {
			g.BallDrawer.Update(screen, true)
		}
	}

	if g.Game.CurrentState == pkg.PlayGame {
		g.BallDrawer.Update(g.Game.Screen, false)
	}

	if g.Game.CurrentState == pkg.PlayGame || g.Game.CurrentState == pkg.ResumeGame {
		if state := g.PlayersDrawer.UpdateLeft(g.Game.Screen); state == pkg.PlayerLLostBall {
			g.playerWinSet(g.Game.PlayerR)
		} else if state := g.PlayersDrawer.UpdateRight(g.Game.Screen); state == pkg.PlayerRLostBall {
			g.playerWinSet(g.Game.PlayerL)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch g.Game.CurrentState {
		case pkg.PlayGame:
			g.Game.CurrentState = pkg.PauseGame
		case pkg.StartGame:
			g.Game.CurrentState = pkg.ResumeGame
			g.Game.StartNewSet()
		case pkg.WinGame:
			g.Game.CurrentState = pkg.StartGame
			g.Game.ResetGame()
		default:
			g.Game.CurrentState = pkg.PlayGame
		}
	}

	return nil
}

func (g *GameDrawer) playerWinSet(player *pkg.Player) {
	g.Game.CurrentState = pkg.ResumeGame
	g.Game.Mark(player)
	g.Game.EndSet(*player)
	if g.Game.Winner() != nil {
		g.Game.CurrentState = pkg.WinGame
	} else {
		g.Game.StartNewSet()
	}
}

// drawBackgroundZone draws the background
func (g *GameDrawer) drawBackgroundZone(screen *ebiten.Image) {
	screen.Fill(g.Game.Screen.AvailableColors["#bg"])

	titleSize := len(g.Game.Title.Text) * g.Game.Title.FontSize
	titleX := float32(GetXCenterPos(g.Game.Screen.Width, g.Game.Title.Text, int(g.Game.Title.FontSize)))
	titleXMargin := 20

	// title
	DrawText(screen, g.Game.Title.Text, g.Game.Title.Font, g.Game.Title.Color,
		pkg.Position{
			X: float32(g.Game.Screen.Width)/2 - float32(titleSize/2),
			Y: g.Game.Screen.YBottom/2 - float32(g.Game.Title.FontSize)/2},
	)

	DrawText(screen, g.Game.Subtitle.Text, g.Game.Subtitle.Font, g.Game.Subtitle.Color,
		pkg.Position{
			X: float32(GetXCenterPos(g.Game.Screen.Width, g.Game.Subtitle.Text, g.Game.Subtitle.FontSize)),
			Y: g.Game.Screen.YBottom/2 + float32(g.Game.Title.FontSize)/2 + 10},
	)

	DrawRectangle(screen,
		int(titleX)-titleXMargin-10, 20,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) - 40, Y: float32(g.Game.Screen.YBottom/2) - 40},
		g.Game.Screen.AvailableColors["#atari1"])

	DrawRectangle(screen,
		g.Game.Screen.Width/2, 20,
		pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(g.Game.Screen.YBottom/2) - 40},
		g.Game.Screen.AvailableColors["#atari1"])

	DrawRectangle(screen,
		20, g.Game.Screen.Height,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) - 40, Y: float32(g.Game.Screen.YBottom/2) - 20},
		g.Game.Screen.AvailableColors["#atari1"])

	DrawRectangle(screen,
		int(titleX)-titleXMargin-40, 20,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) - 10, Y: float32(g.Game.Screen.YBottom/2) - 10},
		g.Game.Screen.AvailableColors["#atari2"])

	DrawRectangle(screen,
		g.Game.Screen.Width/2, 20,
		pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(g.Game.Screen.YBottom/2) - 10},
		g.Game.Screen.AvailableColors["#atari2"])

	DrawRectangle(screen,
		20, g.Game.Screen.Height,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) - 10, Y: float32(g.Game.Screen.YBottom / 2)},
		g.Game.Screen.AvailableColors["#atari2"])

	DrawRectangle(screen,
		int(titleX)-titleXMargin-70, 20,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) + 20, Y: float32(g.Game.Screen.YBottom/2) + 20},
		g.Game.Screen.AvailableColors["#atari3"])

	DrawRectangle(screen,
		g.Game.Screen.Width/2, 20,
		pkg.Position{X: float32(g.Game.Screen.Width/2) + float32(titleSize/2) + 15, Y: float32(g.Game.Screen.YBottom/2) + 20},
		g.Game.Screen.AvailableColors["#atari3"])

	DrawRectangle(screen,
		20, g.Game.Screen.Height,
		pkg.Position{X: float32(g.Game.Screen.XLeft/2) + 20, Y: float32(g.Game.Screen.YBottom/2) + 20},
		g.Game.Screen.AvailableColors["#atari3"])
}

// drawGameZoneTextZone draws the "how to play" and the players's score zone (name and score)
func (g *GameDrawer) drawGameZoneTextZone(screen *ebiten.Image) {
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
			"# THE WIN",
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
			"",
			"press [space] to start or pause",
			"at every moment...")

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
