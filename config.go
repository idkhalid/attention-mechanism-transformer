package main

import (
	"image/color"
	"time"
)

// Application configuration and constants
const (
	ScreenWidth  = 1920
	ScreenHeight = 1080
	TargetFPS    = 60

	WindowTitle = "ATTENTION: The Mathematics of Changing Context"
)

// Scene Identifiers
type SceneID int

const (
	Scene01InitialDistance SceneID = 1
	Scene02Convergence     SceneID = 2
	Scene03NewContext      SceneID = 3
	Scene04Separation       SceneID = 4
	Scene05Aftermath        SceneID = 5
)

// Scene durations in seconds (at 1.0x playback speed)
var SceneDurations = map[SceneID]float64{
	Scene01InitialDistance: 7.0,
	Scene02Convergence:     8.5,
	Scene03NewContext:      8.5,
	Scene04Separation:       9.0,
	Scene05Aftermath:        10.0,
}

// Scene Metadata
type SceneMeta struct {
	ID       SceneID
	Number   string
	Title    string
	Subtitle string
	Concept  string
}

var SceneMetas = map[SceneID]SceneMeta{
	Scene01InitialDistance: {
		ID:       Scene01InitialDistance,
		Number:   "SCENE 01",
		Title:    "Initial Distance",
		Subtitle: "Independent isolated representations in token space",
		Concept:  "Tokens A and B exist as isolated vectors in R^2 with no contextual interaction.",
	},
	Scene02Convergence: {
		ID:       Scene02Convergence,
		Number:   "SCENE 02",
		Title:    "Convergence & Resonance",
		Subtitle: "Pairwise scaled dot-product attention formation",
		Concept:  "Attention scores QK^T / sqrt(d_k) couple the curves: alpha_AA = 0.67, alpha_AB = 0.33.",
	},
	Scene03NewContext: {
		ID:       Scene03NewContext,
		Number:   "SCENE 03",
		Title:    "A New Context",
		Subtitle: "Token C enters — sequence context expands",
		Concept:  "Token C introduces Q_C=[1,1], K_C=[1,1]. The dot-product matrix expands to 3x3.",
	},
	Scene04Separation: {
		ID:       Scene04Separation,
		Number:   "SCENE 04",
		Title:    "Separation & Reallocation",
		Subtitle: "Softmax redistribution of attention energy",
		Concept:  "Attention weights between A and B drop from 0.33 to 0.198, diverted toward Token C (0.401).",
	},
	Scene05Aftermath: {
		ID:       Scene05Aftermath,
		Number:   "SCENE 05",
		Title:    "Aftermath",
		Subtitle: "Contextualized representations in Transformer space",
		Concept:  "The same initial tokens have been transformed into new contextual vectors: O = softmax(S) V.",
	},
}

// Color Palette (Sophisticated Dark Scientific Theme)
var (
	// Backgrounds
	ColorBgDark      = color.RGBA{R: 9, G: 12, B: 18, A: 255}       // #090C12
	ColorBgPanel     = color.RGBA{R: 16, G: 22, B: 34, A: 220}     // #101622
	ColorBgPanelDark = color.RGBA{R: 12, G: 16, B: 26, A: 240}     // #0C101A
	ColorBorder      = color.RGBA{R: 38, G: 50, B: 75, A: 180}     // #26324B
	ColorGrid        = color.RGBA{R: 26, G: 34, B: 52, A: 100}     // Faint coordinate grid
	ColorGridSubtle  = color.RGBA{R: 18, G: 24, B: 38, A: 60}

	// Token A (Cyan / Ice Blue)
	ColorTokenA      = color.RGBA{R: 0, G: 229, B: 255, A: 255}    // #00E5FF
	ColorTokenAGlow  = color.RGBA{R: 0, G: 229, B: 255, A: 60}
	ColorTokenADim   = color.RGBA{R: 0, G: 160, B: 190, A: 180}

	// Token B (Amber / Warm Gold)
	ColorTokenB      = color.RGBA{R: 255, G: 179, B: 0, A: 255}    // #FFB300
	ColorTokenBGlow  = color.RGBA{R: 255, G: 179, B: 0, A: 60}
	ColorTokenBDim   = color.RGBA{R: 190, G: 130, B: 0, A: 180}

	// Token C (Electric Magenta / Violet)
	ColorTokenC      = color.RGBA{R: 224, G: 64, B: 251, A: 255}   // #E040FB
	ColorTokenCGlow  = color.RGBA{R: 224, G: 64, B: 251, A: 60}
	ColorTokenCDim   = color.RGBA{R: 170, G: 45, B: 195, A: 180}

	// Typography & Accents
	ColorTextPrimary   = color.RGBA{R: 240, G: 244, B: 250, A: 255}
	ColorTextSecondary = color.RGBA{R: 150, G: 165, B: 190, A: 255}
	ColorTextMuted     = color.RGBA{R: 90, G: 105, B: 130, A: 255}
	ColorAccent        = color.RGBA{R: 120, G: 180, B: 255, A: 255}
	ColorAccentGreen   = color.RGBA{R: 0, G: 230, B: 118, A: 255}

	// Attention Filaments
	ColorFilamentAB    = color.RGBA{R: 140, G: 210, B: 180, A: 140}
	ColorFilamentAC    = color.RGBA{R: 160, G: 170, B: 255, A: 160}
	ColorFilamentBC    = color.RGBA{R: 240, G: 140, B: 180, A: 160}
)

// UI and layout dimensions
const (
	HUDWidth        = 440
	HUDMargin       = 28
	HeaderHeight    = 70
	FooterHeight    = 65
	CurveCanvasYMin = 90
	CurveCanvasYMax = 990
)

// Tick duration helper
func FrameDuration() time.Duration {
	return time.Second / TargetFPS
}

