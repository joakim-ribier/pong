package pkg

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"golang.org/x/text/language"
)

type Game struct {
	*GameState

	GameMode GameMode

	Title    FontText
	Subtitle FontText
	Screen   Screen

	PlayerL *Player
	PlayerR *Player
	Ball    *Ball

	Win Win

	Debug bool
}

type Win struct {
	Sets         []*Set
	Score        int
	SetScore     int
	SetGapWScore int
}

type Set struct {
	StartTime     time.Time
	EndTime       time.Time
	PlayerLScore  int
	PlayerRScore  int
	PlayerSideWin PlayerSide
	XSpeed        float32
	NbHit         int
}

func (s Set) HasWin(player Player) bool {
	return player.Side == s.PlayerSideWin
}

func (s Set) Duration() string {
	duration := s.EndTime.Sub(s.StartTime)
	return time.Unix(0, 0).UTC().Add(duration).Format("04:05")
}

func (s Set) Speed() float32 {
	return genericsutil.OrElse[float32](
		s.XSpeed, func(v float32) bool { return v > 0 }, func() float32 { return s.XSpeed * -1 })
}

func (s Set) SpeedFormat() string {
	speed := s.Speed()
	if s.Speed() < 10 {
		return fmt.Sprintf("0%0.02f", speed)
	} else {
		return fmt.Sprintf("%0.02f", speed)
	}
}

type FontText struct {
	Text     string
	Font     text.Face
	FontSize int
	Color    color.Color
}

type GameState struct {
	*ResumeGameState

	CurrentState State
	Reset        Reset
}

type ResumeGameState struct {
	Max   int
	Count int
}

type Reset struct {
	Ball Ball
}

func NewGame(mode GameMode, debug bool) *Game {
	remoteExtendZoneW := 0
	if mode != LocalMode {
		remoteExtendZoneW = 350
	}

	screen := Screen{
		Width: 1100 + remoteExtendZoneW, Height: 800, RemoteExtendZoneW: remoteExtendZoneW,
		XLeft: 110 + float32(remoteExtendZoneW), XRight: 1090 + float32(remoteExtendZoneW), YBottom: 110, YTop: 790,
		AvailableColors: map[string]color.Color{
			"#atari1":   color.RGBA{246, 125, 34, 255},
			"#atari2":   color.RGBA{249, 162, 29, 255},
			"#atari3":   color.RGBA{253, 220, 57, 255},
			"#bg":       color.RGBA{77, 77, 77, 255},
			"#playerL":  color.White,
			"#playerR":  color.White,
			"#table-bg": color.RGBA{246, 125, 34, 255},
			"#title":    color.RGBA{120, 226, 160, 255},
			"white":     color.White,
			"black":     color.White},
		Font: NewFont()}

	ball := NewBall(16, 16, Position{
		X: float32(screen.GameZoneXCenter()) - 8,
		Y: float32(screen.GameZoneYCenter()) - 8},
	)

	return &Game{
		GameMode: mode,
		Title: FontText{
			Text: "PONG",
			Font: screen.Font.H1, FontSize: screen.Font.H1Size,
			Color: screen.AvailableColors["white"],
		},
		Subtitle: FontText{
			Text: "joakim-ribier/pong",
			Font: screen.Font.TinyText, FontSize: screen.Font.TinyTextSize,
			Color: screen.AvailableColors["white"],
		},
		Debug:  debug,
		Screen: screen,
		PlayerL: NewPlayer(
			"Player L",
			PlayerLeft,
			NewPaddle(15, 100, Position{
				X: float32(screen.XLeft),
				Y: float32(screen.GameZoneYCenter()) - 50,
			}),
			Options{
				Color: screen.AvailableColors["#playerL"],
				Up:    ebiten.KeyW,
				Down:  ebiten.KeyS}),
		PlayerR: NewPlayer(
			"Player R",
			PlayerRight,
			NewPaddle(15, 100, Position{
				X: float32(screen.XRight - 15),
				Y: float32(screen.GameZoneYCenter()) - 50,
			}),
			Options{
				Color: screen.AvailableColors["#playerR"],
				Up:    ebiten.KeyUp,
				Down:  ebiten.KeyDown}),
		Ball: ball,
		GameState: &GameState{
			CurrentState:    StartGame,
			ResumeGameState: &ResumeGameState{Max: 3, Count: 0},
			Reset:           Reset{Ball: *ball},
		},
		Win: Win{Sets: nil, Score: 11, SetScore: 3, SetGapWScore: 2},
	}
}

type Screen struct {
	Width, Height, RemoteExtendZoneW int
	XLeft, XRight, YBottom, YTop     float32
	AvailableColors                  map[string]color.Color
	Font                             Font
}

func (s Screen) GameZoneWidth() int {
	return (int(s.XRight) - int(s.XLeft))
}

func (s Screen) GameZoneXCenter() int {
	return (s.GameZoneWidth() / 2) + int(s.XLeft)
}

func (s Screen) GameZoneHeight() int {
	return (int(s.YTop) - int(s.YBottom))
}

func (s Screen) GameZoneYCenter() int {
	return (s.GameZoneHeight() / 2) + int(s.YBottom)
}

type Font struct {
	H1, H2, Text, SmallText, TinyText                     text.Face
	H1Size, H2Size, TextSize, SmallTextSize, TinyTextSize int
	AvailableFonts                                        map[string]*text.GoTextFaceSource
}

func NewFont() Font {
	tinyTextSize := 7
	smallTextSize := 8
	textSize := 12
	h2Size := textSize * 2
	h1Size := textSize * 3

	regularFontFace, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}

	availableFonts := map[string]*text.GoTextFaceSource{"#regular": regularFontFace}

	h1Font := &text.GoTextFace{
		Source:    availableFonts["#regular"],
		Direction: text.DirectionLeftToRight,
		Size:      float64(h1Size),
		Language:  language.English,
	}

	h2Font := &text.GoTextFace{
		Source:    availableFonts["#regular"],
		Direction: text.DirectionLeftToRight,
		Size:      float64(h2Size),
		Language:  language.English,
	}

	textFont := &text.GoTextFace{
		Source:    availableFonts["#regular"],
		Direction: text.DirectionLeftToRight,
		Size:      float64(textSize),
		Language:  language.English,
	}
	smallTextFont := &text.GoTextFace{
		Source:    availableFonts["#regular"],
		Direction: text.DirectionLeftToRight,
		Size:      float64(smallTextSize),
		Language:  language.English,
	}
	tinyTextFont := &text.GoTextFace{
		Source:    availableFonts["#regular"],
		Direction: text.DirectionLeftToRight,
		Size:      float64(tinyTextSize),
		Language:  language.English,
	}

	return Font{
		H1: h1Font, H2: h2Font, Text: textFont, SmallText: smallTextFont, TinyText: tinyTextFont,
		H1Size: h1Size, H2Size: h2Size, TextSize: textSize, SmallTextSize: smallTextSize, TinyTextSize: tinyTextSize,
		AvailableFonts: availableFonts,
	}
}

type State int

const (
	PauseGame State = iota
	PlayerLLostBall
	PlayerRLostBall
	PlayGame
	ResumeGame
	StartGame
	WinGame
)

func (s State) String() string {
	switch s {
	case PauseGame:
		return "PauseGame"
	case PlayerLLostBall:
		return "PlayerLLostBall"
	case PlayerRLostBall:
		return "PlayerRLostBall"
	case PlayGame:
		return "PlayGame"
	case ResumeGame:
		return "ResumeGame"
	case StartGame:
		return "StartGame"
	case WinGame:
		return "WinGame"
	default:
		return "Unknown"
	}
}

func ToState(s string) State {
	switch s {
	case "PauseGame":
		return PauseGame
	case "PlayerLLostBall":
		return PlayerLLostBall
	case "PlayerRLostBall":
		return PlayerRLostBall
	case "PlayGame":
		return PlayGame
	case "ResumeGame":
		return ResumeGame
	case "StartGame":
		return StartGame
	case "WinGame":
		return WinGame
	default:
		return -1
	}
}

func (s State) PlayerLostBall() bool {
	return s == PlayerLLostBall || s == PlayerRLostBall
}

func (g Game) IsRemoteServer() bool {
	return g.GameMode == RemoteServerMode
}

func (g Game) IsRemoteClient() bool {
	return g.GameMode == RemoteClientMode
}

func (g Game) IsLocal() bool {
	return !g.IsRemoteServer() && !g.IsRemoteClient()
}

// MaxXSpeedSet returns the max x speed of the sets
func (g Game) MaxXSpeedSet() float32 {
	maxXSpeed := float32(0)
	for _, set := range g.Win.Sets {
		if set.Speed() > maxXSpeed {
			maxXSpeed = set.Speed()
		}
	}
	return maxXSpeed
}

// findTheOtherOne returns the player R if the {playerSide} value is Left otherwise the player L
func (g Game) findTheOtherOne(player Player) *Player {
	return genericsutil.When[Player, *Player](player, func(p Player) bool {
		return p.Side == PlayerLeft
	}, func(p Player) *Player { return g.PlayerR }, func() *Player { return g.PlayerL })
}

// Mark increments the player's winner score and the another player to {win=false}
func (g Game) Mark(player *Player) {
	player.Mark()
	g.findTheOtherOne(*player).Win = false
}

// Winner gets the winner if there is one
func (g Game) Winner() *Player {
	if g.PlayerL.Score >= g.Win.SetScore || g.PlayerR.Score >= g.Win.SetScore {
		if g.PlayerL.Score-g.PlayerR.Score >= g.Win.SetGapWScore {
			return g.PlayerL
		} else if g.PlayerL.Score-g.PlayerR.Score <= -g.Win.SetGapWScore {
			return g.PlayerR
		}
	}

	if g.PlayerL.Score+g.PlayerR.Score == g.Win.Score {
		return genericsutil.OrElse[*Player](g.PlayerL,
			func(p *Player) bool { return p.Score > g.PlayerR.Score }, func() *Player { return g.PlayerR })
	}

	return nil
}

// Looser gets the looser of the game
func (g Game) Looser() *Player {
	if player := g.Winner(); player != nil {
		return g.findTheOtherOne(*player)
	}
	return nil
}

// StartNewSet initializes a new set
func (g *Game) StartNewSet() {
	g.Ball.Position = g.GameState.Reset.Ball.Position
	g.Ball.Impressions = g.GameState.Reset.Ball.Impressions
	g.Ball.XSpeed = g.GameState.Reset.Ball.XSpeed
	g.Ball.YSpeed = g.GameState.Reset.Ball.YSpeed
	g.ResumeGameState.Count = 0

	g.Win.Sets = append(g.Win.Sets, &Set{
		StartTime:    time.Now(),
		XSpeed:       g.Ball.XSpeed,
		PlayerLScore: 0,
		PlayerRScore: 0})
}

// EndSet updates parameters of the current set
func (g *Game) EndSet(player Player) {
	if len(g.Win.Sets) > 0 && g.Win.Sets[len(g.Win.Sets)-1].EndTime.IsZero() {
		currentSet := g.Win.Sets[len(g.Win.Sets)-1]
		currentSet.EndTime = time.Now()
		currentSet.PlayerSideWin = player.Side
		currentSet.PlayerLScore = g.PlayerL.Score
		currentSet.PlayerRScore = g.PlayerR.Score
	}
}

// EndSet updates parameters of the current set
func (g *Game) LastEndedSet() *Set {
	winSets := slicesutil.FilterT[*Set](g.Win.Sets, func(s *Set) bool { return !s.EndTime.IsZero() })
	if len(winSets) == 0 {
		return nil
	}
	return g.Win.Sets[len(winSets)-1]
}

// Hit computes the nb hits of the current set
func (g *Game) Hit() {
	if set := g.CurrentSet(); set != nil {
		set.NbHit += 1
	}
}

// SetXSpeed sets the X speed of the current set
func (g *Game) SetXSpeed(xSpeed float32) {
	if set := g.CurrentSet(); set != nil {
		set.XSpeed = genericsutil.OrElse[float32](xSpeed, func(v float32) bool { return v > 0 }, func() float32 { return xSpeed * -1 })
	}
}

// CurrentSet gets the current set
func (g *Game) CurrentSet() *Set {
	if len(g.Win.Sets) > 0 && g.Win.Sets[len(g.Win.Sets)-1].EndTime.IsZero() {
		return g.Win.Sets[len(g.Win.Sets)-1]
	} else {
		return nil
	}
}

// ResetGame initializes a new Game
func (g *Game) ResetGame() {
	g.PlayerL.Score = 0
	g.PlayerR.Score = 0
	g.Win.Sets = nil
}

type GameMode int

const (
	LocalMode GameMode = iota
	RemoteClientMode
	RemoteServerMode
)
