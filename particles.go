package main

import (
	"image/color"
	"math"
	"math/rand"
)

// ParticleKind classifies particles
type ParticleKind int

const (
	ParticleKindCurveTrail ParticleKind = iota // Rides on token wave
	ParticleKindAttention                      // Flows along attention filaments between tokens
	ParticleKindAmbient                        // Subtle background dust
)

// Particle represents an individual glowing light point
type Particle struct {
	X         float64
	Y         float64
	Vx        float64
	Vy        float64
	Size      float64
	Alpha     float64
	MaxAlpha  float64
	Life      float64
	MaxLife   float64
	Color     color.RGBA
	Kind      ParticleKind
	TokenID   int     // Associated token or source token
	TargetID  int     // Target token for attention particle
	ParamT    float64 // Curve or filament parameter [0.0 - 1.0]
	Speed     float64
}

// ParticleSystem manages lifecycle and emission of particles
type ParticleSystem struct {
	particles []*Particle
	rng       *rand.Rand
	maxCount  int
}

// NewParticleSystem creates a particle manager
func NewParticleSystem(maxCount int) *ParticleSystem {
	ps := &ParticleSystem{
		particles: make([]*Particle, 0, maxCount),
		rng:       rand.New(rand.NewSource(42)),
		maxCount:  maxCount,
	}

	// Initialize ambient background particles
	for i := 0; i < 90; i++ {
		ps.particles = append(ps.particles, &Particle{
			X:        ps.rng.Float64() * ScreenWidth,
			Y:        CurveCanvasYMin + ps.rng.Float64()*(CurveCanvasYMax-CurveCanvasYMin),
			Vx:       (ps.rng.Float64() - 0.5) * 0.25,
			Vy:       (ps.rng.Float64() - 0.5) * 0.15,
			Size:     0.8 + ps.rng.Float64()*1.2,
			Alpha:    0.1 + ps.rng.Float64()*0.25,
			MaxAlpha: 0.35,
			Life:     500 + ps.rng.Float64()*1000,
			MaxLife:  1500,
			Color:    color.RGBA{R: 120, G: 160, B: 220, A: 255},
			Kind:     ParticleKindAmbient,
		})
	}

	return ps
}

// EmitCurveParticle spawns a particle that rides a token's wave
func (ps *ParticleSystem) EmitCurveParticle(tokenID int, x, y float64, col color.RGBA) {
	if len(ps.particles) >= ps.maxCount {
		return
	}
	ps.particles = append(ps.particles, &Particle{
		X:        x,
		Y:        y,
		Vx:       0.6 + ps.rng.Float64()*1.2,
		Vy:       (ps.rng.Float64() - 0.5) * 0.4,
		Size:     1.2 + ps.rng.Float64()*1.8,
		Alpha:    0.6 + ps.rng.Float64()*0.4,
		MaxAlpha: 0.9,
		Life:     60 + ps.rng.Float64()*90,
		MaxLife:  150,
		Color:    col,
		Kind:     ParticleKindCurveTrail,
		TokenID:  tokenID,
	})
}

// EmitAttentionParticle spawns a packet of information traveling from Token source to Token target
func (ps *ParticleSystem) EmitAttentionParticle(srcID, tgtID int, p1, p2 Point2D, col color.RGBA, weight float64) {
	if len(ps.particles) >= ps.maxCount {
		return
	}
	ps.particles = append(ps.particles, &Particle{
		X:        p1.X,
		Y:        p1.Y,
		Size:     1.5 + weight*2.5,
		Alpha:    0.7 + weight*0.3,
		MaxAlpha: 1.0,
		Life:     45,
		MaxLife:  45,
		Color:    col,
		Kind:     ParticleKindAttention,
		TokenID:  srcID,
		TargetID: tgtID,
		ParamT:   0.0,
		Speed:    1.0 / (30.0 + ps.rng.Float64()*20.0),
	})
}

// Update advances the particle physics
func (ps *ParticleSystem) Update(dt float64, curves []*TokenCurve, filaments []*AttentionFilament) {
	alive := ps.particles[:0]

	for _, p := range ps.particles {
		p.Life -= dt * 60.0

		switch p.Kind {
		case ParticleKindAmbient:
			p.X += p.Vx * dt * 60.0
			p.Y += p.Vy * dt * 60.0
			if p.X < 0 {
				p.X = ScreenWidth
			}
			if p.X > ScreenWidth {
				p.X = 0
			}
			if p.Y < CurveCanvasYMin {
				p.Y = CurveCanvasYMax
			}
			if p.Y > CurveCanvasYMax {
				p.Y = CurveCanvasYMin
			}
			// Keep ambient alive
			if p.Life <= 0 {
				p.Life = p.MaxLife
			}
			alive = append(alive, p)

		case ParticleKindCurveTrail:
			if p.Life > 0 {
				p.X += p.Vx * dt * 60.0
				p.Y += p.Vy * dt * 60.0
				progress := p.Life / p.MaxLife
				p.Alpha = p.MaxAlpha * math.Sin(progress*math.Pi)
				alive = append(alive, p)
			}

		case ParticleKindAttention:
			if p.Life > 0 {
				p.ParamT += p.Speed * dt * 60.0
				if p.ParamT > 1.0 {
					p.ParamT = 1.0
					p.Life = 0
				} else {
					// Interpolate position along filament arc
					t := p.ParamT
					// Find matching filament or endpoints
					var p1, p2 Point2D
					found := false
					for _, fil := range filaments {
						if (fil.Token1ID == p.TokenID && fil.Token2ID == p.TargetID) ||
							(fil.Token2ID == p.TokenID && fil.Token1ID == p.TargetID) {
							p1 = fil.P1
							p2 = fil.P2
							found = true
							break
						}
					}
					if !found && p.TokenID < len(curves) && p.TargetID < len(curves) {
						c1 := curves[p.TokenID]
						c2 := curves[p.TargetID]
						p1 = Point2D{X: ScreenWidth * 0.45, Y: c1.CenterY}
						p2 = Point2D{X: ScreenWidth * 0.55, Y: c2.CenterY}
						found = true
					}
					if found {
						// Subtle quadratic Bézier curve arc
						midX := (p1.X+p2.X)*0.5 + 40.0*math.Sin(t*math.Pi)
						midY := (p1.Y + p2.Y) * 0.5
						oneMinusT := 1.0 - t
						p.X = oneMinusT*oneMinusT*p1.X + 2*oneMinusT*t*midX + t*t*p2.X
						p.Y = oneMinusT*oneMinusT*p1.Y + 2*oneMinusT*t*midY + t*t*p2.Y
					}
					p.Alpha = p.MaxAlpha * math.Sin(t*math.Pi)
					alive = append(alive, p)
				}
			}
		}
	}

	ps.particles = alive
}

