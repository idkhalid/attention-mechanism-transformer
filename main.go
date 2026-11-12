package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game implements the ebiten.Game interface
type Game struct {
	simulation *Simulation
	animation  *AnimationManager
	renderer   *Renderer
}

// NewGame initializes the visualization system
func NewGame() (*Game, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize renderer: %w", err)
	}

	sim := NewSimulation()
	anim := NewAnimationManager()

	return &Game{
		simulation: sim,
		animation:  anim,
		renderer:   renderer,
	}, nil
}

// Update handles logic tick (60 FPS)
func (g *Game) Update() error {
	g.handleInput()

	dt := 1.0 / float64(TargetFPS)
	g.animation.Update(dt)
	g.simulation.Update(dt, g.animation)

	return nil
}

// Draw renders the visual scene
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Render(screen, g.simulation, g.animation)
}

// Layout maintains the 1920x1080 logical resolution
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// handleInput dispatches user keyboard and mouse commands
func (g *Game) handleInput() {
	// Play / Pause
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.animation.TogglePause()
	}

	// Restart
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.animation.Restart()
	}

	// Scene Step navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.animation.PrevScene()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.animation.NextScene()
	}

	// Direct scene jump (1 - 5)
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.animation.SetScene(Scene01InitialDistance)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.animation.SetScene(Scene02Convergence)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.animation.SetScene(Scene03NewContext)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.animation.SetScene(Scene04Separation)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		g.animation.SetScene(Scene05Aftermath)
	}

	// Playback speed adjustment
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyRightBracket) {
		g.animation.AdjustSpeed(0.25)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyLeftBracket) {
		g.animation.AdjustSpeed(-0.25)
	}

	// Toggle HUD (Mathematical values)
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.animation.ShowHUD = !g.animation.ShowHUD
	}

	// Toggle Causal Masking (PDF Section 02)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.animation.CausalMask = !g.animation.CausalMask
	}

	// Toggle Help Overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.animation.ShowHelp = !g.animation.ShowHelp
	}

	// Escape key to dismiss Help
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.animation.ShowHelp = false
	}

	// Mouse click handling for timeline scrubber
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		// Check if click is on footer timeline scrubber
		footerY := ScreenHeight - FooterHeight
		if my >= footerY && my <= footerY+30 {
			barX := 40
			barW := ScreenWidth - 80
			if mx >= barX && mx <= barX+barW {
				clickRatio := float64(mx-barX) / float64(barW)
				targetScene := SceneID(int(clickRatio*5.0) + 1)
				if targetScene >= Scene01InitialDistance && targetScene <= Scene05Aftermath {
					g.animation.SetScene(targetScene)
				}
			}
		}
	}
}

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle(WindowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(TargetFPS)

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Critical: failed to start game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Game loop terminated: %v", err)
	}
}

