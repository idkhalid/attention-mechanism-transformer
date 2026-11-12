package main

import (
	"image/color"
	"math"
)

// Point2D represents a 2D coordinate on screen
type Point2D struct {
	X float64
	Y float64
}

// TokenCurve represents the continuous mathematical waveform of a token
type TokenCurve struct {
	ID          int     // 0: Token A, 1: Token B, 2: Token C
	Name        string  // "TOKEN A", "TOKEN B", "TOKEN C"
	BaseColor   color.RGBA
	GlowColor   color.RGBA
	BaseFreq    float64 // Spatial frequency k
	Speed       float64 // Temporal angular velocity omega
	BaseAmp     float64 // Peak amplitude
	Phase       float64 // Initial phase shift

	// Dynamic state computed by simulation
	CenterY     float64 // Current vertical baseline position
	TargetY     float64 // Goal vertical position
	Alpha       float64 // Opacity [0.0 - 1.0]
	Resonance   float64 // Coupling resonance intensity [0.0 - 1.0]

	// Sample points along the screen width
	Points      []Point2D
}

// NewTokenCurve creates an initialized curve
func NewTokenCurve(id int, name string, baseColor, glowColor color.RGBA, freq, speed, amp, phase, centerY float64) *TokenCurve {
	return &TokenCurve{
		ID:        id,
		Name:      name,
		BaseColor: baseColor,
		GlowColor: glowColor,
		BaseFreq:  freq,
		Speed:     speed,
		BaseAmp:   amp,
		Phase:     phase,
		CenterY:   centerY,
		TargetY:   centerY,
		Alpha:     1.0,
		Resonance: 0.0,
		Points:    make([]Point2D, 200),
	}
}

// EvaluateWave calculates the waveform height for token i at position x and time t,
// taking into account attention weights alpha_ij and value vector transfer
func (tc *TokenCurve) EvaluateWave(x, t float64, weights []float64, otherCurves []*TokenCurve) float64 {
	// 1. Intrinsic base wave for this token
	k := tc.BaseFreq
	w := tc.Speed
	phi := tc.Phase

	var wave float64
	switch tc.ID {
	case 0: // Token A: clean fundamental harmonic with subtle secondary overtone
		wave = math.Sin(k*x-w*t+phi) + 0.30*math.Sin(2.0*k*x-1.4*w*t)
	case 1: // Token B: energetic phase-shifted dual harmonic
		wave = math.Sin(k*x-w*t+phi) + 0.35*math.Cos(1.8*k*x-0.8*w*t)
	case 2: // Token C: dense multi-component context wave
		wave = math.Sin(k*x-w*t+phi) + 0.45*math.Sin(2.2*k*x-1.8*w*t+0.5) + 0.2*math.Cos(3.0*k*x)
	default:
		wave = math.Sin(k*x - w*t + phi)
	}

	totalWave := wave * tc.BaseAmp

	// 2. Contextual Transfer: If attention weights are provided,
	// inject harmonics from other tokens weighted by alpha_ij
	if len(weights) > tc.ID && len(otherCurves) > 0 {
		for j, other := range otherCurves {
			if j != tc.ID && j < len(weights) {
				w_ij := weights[j]
				if w_ij > 0.01 && other.Alpha > 0.01 {
					// Other token's intrinsic wave
					otherK := other.BaseFreq
					otherW := other.Speed
					var otherBase float64
					if other.ID == 0 {
						otherBase = math.Sin(otherK*x - otherW*t + other.Phase)
					} else if other.ID == 1 {
						otherBase = math.Sin(otherK*x - otherW*t + other.Phase)
					} else {
						otherBase = math.Sin(otherK*x - otherW*t + other.Phase)
					}

					// Transfer wave harmonic scaled by attention weight alpha_ij
					// and resonance factor
					transferStrength := w_ij * 0.75 * other.Alpha
					totalWave += otherBase * other.BaseAmp * transferStrength
				}
			}
		}
	}

	return totalWave
}

// AttentionFilament represents an energy bond between two tokens
type AttentionFilament struct {
	Token1ID   int
	Token2ID   int
	P1         Point2D
	P2         Point2D
	Weight     float64
	Color      color.RGBA
	Active     bool
	PulsePhase float64
}

