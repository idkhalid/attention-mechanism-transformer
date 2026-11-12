package main

import (
	"math"
)

// Simulation coordinates the mathematical engine, curves, filaments, and particles
type Simulation struct {
	AttentionEngine *AttentionEngine
	Curves          []*TokenCurve
	Filaments       []*AttentionFilament
	Particles       *ParticleSystem
	CurrentResult   AttentionResult

	// Simulation time tracking
	Time float64

	// Contextual representation coordinates for 2D subspace view
	ContextualPoints []Point2D
	OriginalPoints   []Point2D
}

// NewSimulation constructs the simulation
func NewSimulation() *Simulation {
	ae := NewAttentionEngine()

	// Initialize 3 curves: Token A, Token B, Token C
	// Parameters: ID, Name, BaseColor, GlowColor, BaseFreq, Speed, BaseAmp, Phase, CenterY
	curveA := NewTokenCurve(0, "TOKEN A", ColorTokenA, ColorTokenAGlow, 0.0075, 1.8, 38.0, 0.0, 360.0)
	curveB := NewTokenCurve(1, "TOKEN B", ColorTokenB, ColorTokenBGlow, 0.0125, 2.4, 34.0, math.Pi*0.4, 720.0)
	curveC := NewTokenCurve(2, "TOKEN C", ColorTokenC, ColorTokenCGlow, 0.0160, 2.0, 42.0, math.Pi*0.8, 180.0)
	curveC.Alpha = 0.0 // Initially invisible

	curves := []*TokenCurve{curveA, curveB, curveC}

	// Filaments linking pairs: (A-B), (A-C), (B-C)
	filaments := []*AttentionFilament{
		{Token1ID: 0, Token2ID: 1, Color: ColorFilamentAB, Active: false},
		{Token1ID: 0, Token2ID: 2, Color: ColorFilamentAC, Active: false},
		{Token1ID: 1, Token2ID: 2, Color: ColorFilamentBC, Active: false},
	}

	sim := &Simulation{
		AttentionEngine:  ae,
		Curves:           curves,
		Filaments:        filaments,
		Particles:        NewParticleSystem(350),
		CurrentResult:    ae.ComputeScaledDotProduct(2, false),
		Time:             0.0,
		ContextualPoints: make([]Point2D, 3),
		OriginalPoints:   make([]Point2D, 3),
	}

	return sim
}

// Update computes the mathematical state for the current frame
func (s *Simulation) Update(dt float64, anim *AnimationManager) {
	if anim.IsPaused {
		return
	}

	stepDT := dt * anim.PlaybackSpeed
	s.Time += stepDT

	prog := anim.SceneProgress()
	causal := anim.CausalMask

	curveA := s.Curves[0]
	curveB := s.Curves[1]
	curveC := s.Curves[2]

	var tokenCPresence float64

	switch anim.CurrentScene {
	case Scene01InitialDistance:
		// Scene 1: Independent tokens, far apart, no mutual attention
		tokenCPresence = 0.0
		curveC.Alpha = 0.0

		// Target positions: Far apart
		curveA.TargetY = 360.0
		curveB.TargetY = 720.0
		curveC.TargetY = 160.0

		// Amplitudes and resonance
		curveA.Resonance = 0.0
		curveB.Resonance = 0.0
		curveC.Resonance = 0.0

		// Filaments inactive
		for _, f := range s.Filaments {
			f.Active = false
			f.Weight = 0.0
		}

		s.CurrentResult = s.AttentionEngine.ComputeScaledDotProduct(2, causal)

	case Scene02Convergence:
		// Scene 2: Curves converge, attention forms, resonance builds
		tokenCPresence = 0.0
		curveC.Alpha = 0.0

		// Interpolate curves toward central meeting zone
		ease := EaseInOutCubic(prog)
		curveA.TargetY = Lerp(360.0, 500.0, ease)
		curveB.TargetY = Lerp(720.0, 580.0, ease)
		curveC.TargetY = 160.0

		// Attention between A and B
		s.CurrentResult = s.AttentionEngine.ComputeScaledDotProduct(2, causal)

		// Resonance increases
		res := Smoothstep(0.15, 0.85, prog)
		curveA.Resonance = res
		curveB.Resonance = res

		// Filament A-B active with growing strength
		fAB := s.Filaments[0]
		fAB.Active = true
		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 1 {
			fAB.Weight = s.CurrentResult.Weights[0][1] * res
		}

		s.Filaments[1].Active = false
		s.Filaments[1].Weight = 0.0
		s.Filaments[2].Active = false
		s.Filaments[2].Weight = 0.0

	case Scene03NewContext:
		// Scene 3: Token C enters unexpectedly! Softmax begins redistribution
		tokenCPresence = EaseInOutQuad(prog)
		curveC.Alpha = Smoothstep(0.05, 0.40, prog)

		// Token C descends to center stage
		curveC.TargetY = Lerp(160.0, 540.0, EaseInOutCubic(prog))

		// Token A and B start reacting to C's presence
		curveA.TargetY = Lerp(500.0, 440.0, EaseInOutQuad(prog))
		curveB.TargetY = Lerp(580.0, 640.0, EaseInOutQuad(prog))

		s.CurrentResult = s.AttentionEngine.InterpolatedAttention(tokenCPresence, causal)

		// Update filaments
		fAB := s.Filaments[0]
		fAC := s.Filaments[1]
		fBC := s.Filaments[2]

		fAB.Active = true
		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 1 {
			fAB.Weight = s.CurrentResult.Weights[0][1] // Decreasing!
		}

		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 2 {
			fAC.Active = prog > 0.15 && s.CurrentResult.Weights[0][2] > 0.01
			fAC.Weight = s.CurrentResult.Weights[0][2] // Increasing!
		} else {
			fAC.Active = false
			fAC.Weight = 0.0
		}

		if len(s.CurrentResult.Weights) > 1 && len(s.CurrentResult.Weights[1]) > 2 {
			fBC.Active = prog > 0.15 && s.CurrentResult.Weights[1][2] > 0.01
			fBC.Weight = s.CurrentResult.Weights[1][2] // Increasing!
		} else {
			fBC.Active = false
			fBC.Weight = 0.0
		}

		curveA.Resonance = Lerp(1.0, 0.3, prog)
		curveB.Resonance = Lerp(1.0, 0.3, prog)
		curveC.Resonance = prog

	case Scene04Separation:
		// Scene 4: Token A and B pushed apart by context redistribution
		tokenCPresence = 1.0
		curveC.Alpha = 1.0

		ease := EaseInOutCubic(prog)
		curveA.TargetY = Lerp(440.0, 310.0, ease)
		curveB.TargetY = Lerp(640.0, 770.0, ease)
		curveC.TargetY = 540.0

		s.CurrentResult = s.AttentionEngine.ComputeScaledDotProduct(3, causal)

		// Filaments: A-B is now weak (0.198), while A-C and B-C are dominant (0.401)
		s.Filaments[0].Active = true
		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 1 {
			s.Filaments[0].Weight = s.CurrentResult.Weights[0][1]
		}

		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 2 {
			s.Filaments[1].Active = true
			s.Filaments[1].Weight = s.CurrentResult.Weights[0][2]
		}

		if len(s.CurrentResult.Weights) > 1 && len(s.CurrentResult.Weights[1]) > 2 {
			s.Filaments[2].Active = true
			s.Filaments[2].Weight = s.CurrentResult.Weights[1][2]
		}

		curveA.Resonance = 0.2
		curveB.Resonance = 0.2
		curveC.Resonance = 0.8

	case Scene05Aftermath:
		// Scene 5: Stable contextual representations
		tokenCPresence = 1.0
		curveC.Alpha = 1.0

		curveA.TargetY = 310.0
		curveB.TargetY = 770.0
		curveC.TargetY = 540.0

		s.CurrentResult = s.AttentionEngine.ComputeScaledDotProduct(3, causal)

		for i := 0; i < 3; i++ {
			s.Curves[i].Resonance = 0.5
		}

		s.Filaments[0].Active = true
		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 1 {
			s.Filaments[0].Weight = s.CurrentResult.Weights[0][1]
		}

		if len(s.CurrentResult.Weights) > 0 && len(s.CurrentResult.Weights[0]) > 2 {
			s.Filaments[1].Active = true
			s.Filaments[1].Weight = s.CurrentResult.Weights[0][2]
		}

		if len(s.CurrentResult.Weights) > 1 && len(s.CurrentResult.Weights[1]) > 2 {
			s.Filaments[2].Active = true
			s.Filaments[2].Weight = s.CurrentResult.Weights[1][2]
		}
	}

	// Smoothly interpolate curve vertical center positions
	lerpSpeed := 6.0 * stepDT
	if lerpSpeed > 1.0 {
		lerpSpeed = 1.0
	}
	for _, c := range s.Curves {
		c.CenterY += (c.TargetY - c.CenterY) * lerpSpeed
	}

	// Sample curve points across the screen canvas
	canvasLeft := 60.0
	canvasRight := ScreenWidth - HUDWidth - 80.0
	numSamples := len(curveA.Points)
	stepX := (canvasRight - canvasLeft) / float64(numSamples-1)

	for i, c := range s.Curves {
		var weights []float64
		if i < len(s.CurrentResult.Weights) {
			weights = s.CurrentResult.Weights[i]
		}
		for idx := 0; idx < numSamples; idx++ {
			x := canvasLeft + float64(idx)*stepX
			waveY := c.EvaluateWave(x, s.Time, weights, s.Curves)
			c.Points[idx] = Point2D{
				X: x,
				Y: c.CenterY + waveY,
			}
		}
	}

	// Update filament anchor points at sample fractions
	midIdx := numSamples / 2
	for _, f := range s.Filaments {
		if f.Active && f.Token1ID < len(s.Curves) && f.Token2ID < len(s.Curves) {
			f.P1 = s.Curves[f.Token1ID].Points[midIdx]
			f.P2 = s.Curves[f.Token2ID].Points[midIdx]
			f.PulsePhase += stepDT * 3.0
		}
	}

	// Particle emissions
	if !anim.IsPaused {
		// Emit wave trail particles
		for i, c := range s.Curves {
			if c.Alpha > 0.05 {
				// Pick sample along wave
				sampleIdx := int(s.Time*30.0+float64(i*45)) % numSamples
				pt := c.Points[sampleIdx]
				s.Particles.EmitCurveParticle(i, pt.X, pt.Y, c.BaseColor)
			}
		}

		// Emit attention filament particles
		for _, f := range s.Filaments {
			if f.Active && f.Weight > 0.05 {
				if s.Particles.rng.Float64() < f.Weight*1.2 {
					s.Particles.EmitAttentionParticle(f.Token1ID, f.Token2ID, f.P1, f.P2, f.Color, f.Weight)
				}
			}
		}
	}

	// Advance particles
	s.Particles.Update(stepDT, s.Curves, s.Filaments)

	// Update 2D embedding subspace coordinates for HUD
	// Initial V vectors: A=[1, 0], B=[0, 2], C=[2, 1]
	// We scale and shift them into an inset plot
	subspaceCenterX := float64(ScreenWidth - HUDWidth + 120)
	subspaceCenterY := 880.0
	scaleSubspace := 42.0

	for i := 0; i < 3; i++ {
		orig := s.AttentionEngine.BaseV[i]
		s.OriginalPoints[i] = Point2D{
			X: subspaceCenterX + orig[0]*scaleSubspace,
			Y: subspaceCenterY - orig[1]*scaleSubspace,
		}

		if i < len(s.CurrentResult.Output) {
			out := s.CurrentResult.Output[i]
			s.ContextualPoints[i] = Point2D{
				X: subspaceCenterX + out[0]*scaleSubspace,
				Y: subspaceCenterY - out[1]*scaleSubspace,
			}
		}
	}
}

