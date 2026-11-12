package main

import (
	"bytes"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
)

// Renderer handles all 2D vector drawing, typography, and mathematical UI
type Renderer struct {
	regularSource *text.GoTextFaceSource
	monoSource    *text.GoTextFaceSource

	// Font faces
	faceHero      *text.GoTextFace
	faceTitle     *text.GoTextFace
	faceHeading   *text.GoTextFace
	faceBody      *text.GoTextFace
	faceSmall     *text.GoTextFace
	faceMono      *text.GoTextFace
	faceMonoSmall *text.GoTextFace

	// Offscreen scratch image for post-processing if needed
	offscreen *ebiten.Image
}

// NewRenderer loads fonts and initializes graphics objects
func NewRenderer() (*Renderer, error) {
	regSrc, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, fmt.Errorf("failed to load regular font: %w", err)
	}

	monoSrc, err := text.NewGoTextFaceSource(bytes.NewReader(gomono.TTF))
	if err != nil {
		return nil, fmt.Errorf("failed to load mono font: %w", err)
	}

	r := &Renderer{
		regularSource: regSrc,
		monoSource:    monoSrc,
		faceHero:      &text.GoTextFace{Source: regSrc, Size: 38},
		faceTitle:     &text.GoTextFace{Source: regSrc, Size: 26},
		faceHeading:   &text.GoTextFace{Source: regSrc, Size: 18},
		faceBody:      &text.GoTextFace{Source: regSrc, Size: 14},
		faceSmall:     &text.GoTextFace{Source: regSrc, Size: 11},
		faceMono:      &text.GoTextFace{Source: monoSrc, Size: 13},
		faceMonoSmall: &text.GoTextFace{Source: monoSrc, Size: 10.5},
		offscreen:     ebiten.NewImage(ScreenWidth, ScreenHeight),
	}

	return r, nil
}

// DrawText draws text at (x, y) with a given face and color
func (r *Renderer) DrawText(dst *ebiten.Image, str string, face *text.GoTextFace, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, str, face, op)
}

// DrawTextAligned draws text with alignment
func (r *Renderer) DrawTextAligned(dst *ebiten.Image, str string, face *text.GoTextFace, x, y float64, clr color.Color, align text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	op.PrimaryAlign = align
	text.Draw(dst, str, face, op)
}

// Render coordinates the complete visual frame
func (r *Renderer) Render(screen *ebiten.Image, sim *Simulation, anim *AnimationManager) {
	// 1. Background clear
	screen.Fill(ColorBgDark)

	// 2. Deep mathematical coordinate grid
	r.drawCoordinateGrid(screen, sim)

	// 3. Ambient and Attention particles
	r.drawParticles(screen, sim.Particles)

	// 4. Attention Filaments (Energy bonds)
	r.drawFilaments(screen, sim)

	// 5. Mathematical curves with multi-pass neon glow
	r.drawCurves(screen, sim)

	// 6. Scene 05 Grand Epilogue Overlay
	if anim.CurrentScene == Scene05Aftermath {
		r.drawAftermathOverlay(screen, anim)
	}

	// 7. Top cinematic title banner & scene descriptions
	r.drawHeader(screen, anim)

	// 8. Mathematical HUD Inspector (Right side panel)
	if anim.ShowHUD {
		r.drawHUD(screen, sim, anim)
	}

	// 9. Timeline scrubber and control footer (Bottom)
	r.drawFooter(screen, anim)

	// 10. Help modal if toggled
	if anim.ShowHelp {
		r.drawHelpOverlay(screen)
	}
}

// drawCoordinateGrid draws a clean, restrained Cartesian coordinate backdrop
func (r *Renderer) drawCoordinateGrid(screen *ebiten.Image, sim *Simulation) {
	gridRight := float32(ScreenWidth - HUDWidth - 40)
	gridLeft := float32(50)
	gridTop := float32(HeaderHeight + 20)
	gridBottom := float32(ScreenHeight - FooterHeight - 20)

	// Vertical grid lines
	stepX := float32(80)
	for x := gridLeft; x <= gridRight; x += stepX {
		vector.StrokeLine(screen, x, gridTop, x, gridBottom, 1.0, ColorGridSubtle, false)
	}

	// Horizontal grid lines
	stepY := float32(70)
	for y := gridTop; y <= gridBottom; y += stepY {
		vector.StrokeLine(screen, gridLeft, y, gridRight, y, 1.0, ColorGridSubtle, false)
	}

	// Primary Center Axis (Y = 540)
	centerY := float32(540)
	vector.StrokeLine(screen, gridLeft, centerY, gridRight, centerY, 1.5, ColorGrid, false)

	// Faint tick labels
	r.DrawText(screen, "Y=0 (Convergence Plane)", r.faceMonoSmall, float64(gridLeft+10), float64(centerY-16), ColorTextMuted)
	r.DrawText(screen, "+y (Token Space A)", r.faceMonoSmall, float64(gridLeft+10), float64(gridTop+14), ColorTextMuted)
	r.DrawText(screen, "-y (Token Space B)", r.faceMonoSmall, float64(gridLeft+10), float64(gridBottom-16), ColorTextMuted)
}

// drawCurves renders the mathematical curves with layered glow
func (r *Renderer) drawCurves(screen *ebiten.Image, sim *Simulation) {
	for _, c := range sim.Curves {
		if c.Alpha <= 0.005 || len(c.Points) < 2 {
			continue
		}

		baseAlpha := c.Alpha

		// Colors with alpha scaling
		glowOuter := color.RGBA{
			R: c.GlowColor.R,
			G: c.GlowColor.G,
			B: c.GlowColor.B,
			A: uint8(float64(c.GlowColor.A) * baseAlpha * 0.4),
		}
		glowInner := color.RGBA{
			R: c.BaseColor.R,
			G: c.BaseColor.G,
			B: c.BaseColor.B,
			A: uint8(100.0 * baseAlpha),
		}
		coreColor := color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: uint8(230.0 * baseAlpha),
		}
		tintColor := color.RGBA{
			R: c.BaseColor.R,
			G: c.BaseColor.G,
			B: c.BaseColor.B,
			A: uint8(255.0 * baseAlpha),
		}

		// Multi-pass polyline drawing
		for p := 0; p < len(c.Points)-1; p++ {
			pt1 := c.Points[p]
			pt2 := c.Points[p+1]

			x0, y0 := float32(pt1.X), float32(pt1.Y)
			x1, y1 := float32(pt2.X), float32(pt2.Y)

			// 1. Wide outer halo pass
			vector.StrokeLine(screen, x0, y0, x1, y1, 7.0, glowOuter, true)
			// 2. Medium bloom pass
			vector.StrokeLine(screen, x0, y0, x1, y1, 3.5, glowInner, true)
			// 3. Tinted body
			vector.StrokeLine(screen, x0, y0, x1, y1, 2.0, tintColor, true)
			// 4. Ultra-fine bright core
			vector.StrokeLine(screen, x0, y0, x1, y1, 1.0, coreColor, true)
		}

		// Head node indicator (Pulsing glowing orb)
		headIdx := len(c.Points) * 3 / 4
		headPt := c.Points[headIdx]
		pulse := 1.0 + 0.2*math.Sin(sim.Time*4.0+float64(c.ID))

		vector.FillCircle(screen, float32(headPt.X), float32(headPt.Y), float32(10.0*pulse), glowOuter, true)
		vector.FillCircle(screen, float32(headPt.X), float32(headPt.Y), float32(5.5*pulse), glowInner, true)
		vector.FillCircle(screen, float32(headPt.X), float32(headPt.Y), 3.0, coreColor, true)

		// Token label tag
		tagX := headPt.X + 16.0
		tagY := headPt.Y - 10.0

		// Badge background
		vector.FillRect(screen, float32(tagX-6), float32(tagY-4), 105, 24, ColorBgPanelDark, true)
		vector.StrokeRect(screen, float32(tagX-6), float32(tagY-4), 105, 24, 1.0, tintColor, false)

		var vectorStr string
		switch c.ID {
		case 0:
			vectorStr = "[1, 0]"
		case 1:
			vectorStr = "[0, 1]"
		case 2:
			vectorStr = "[1, 1]"
		}
		label := fmt.Sprintf("%s %s", c.Name, vectorStr)
		r.DrawText(screen, label, r.faceMonoSmall, tagX, tagY+12, tintColor)
	}
}

// drawFilaments renders attention energy arcs between interacting curves
func (r *Renderer) drawFilaments(screen *ebiten.Image, sim *Simulation) {
	for _, f := range sim.Filaments {
		if !f.Active || f.Weight < 0.01 {
			continue
		}

		// Width and alpha proportional to attention weight
		width := float32(1.5 + f.Weight*5.0)
		alpha := uint8(math.Min(255.0, f.Weight*250.0))

		filColor := color.RGBA{
			R: f.Color.R,
			G: f.Color.G,
			B: f.Color.B,
			A: alpha,
		}

		glowColor := color.RGBA{
			R: f.Color.R,
			G: f.Color.G,
			B: f.Color.B,
			A: uint8(float64(alpha) * 0.35),
		}

		// Render dotted/segmented energy beam
		numSegments := 20
		p1 := f.P1
		p2 := f.P2

		for s := 0; s < numSegments; s++ {
			t0 := float64(s) / float64(numSegments)
			t1 := float64(s+1) / float64(numSegments)

			// Subtle curve deflection based on pulse phase
			arcFactor := math.Sin(t0*math.Pi) * 22.0 * math.Sin(f.PulsePhase)

			x0 := Lerp(p1.X, p2.X, t0) + arcFactor
			y0 := Lerp(p1.Y, p2.Y, t0)
			x1 := Lerp(p1.X, p2.X, t1) + math.Sin(t1*math.Pi)*22.0*math.Sin(f.PulsePhase)
			y1 := Lerp(p1.Y, p2.Y, t1)

			// Glow line
			vector.StrokeLine(screen, float32(x0), float32(y0), float32(x1), float32(y1), width*2.2, glowColor, true)
			// Core filament
			vector.StrokeLine(screen, float32(x0), float32(y0), float32(x1), float32(y1), width, filColor, true)
		}

		// Midpoint badge showing alpha_ij weight
		midX := (p1.X + p2.X) * 0.5
		midY := (p1.Y + p2.Y) * 0.5
		badgeText := fmt.Sprintf("α = %.3f", f.Weight)

		vector.FillRect(screen, float32(midX-36), float32(midY-11), 72, 22, ColorBgPanelDark, true)
		vector.StrokeRect(screen, float32(midX-36), float32(midY-11), 72, 22, 1.0, f.Color, false)
		r.DrawTextAligned(screen, badgeText, r.faceMonoSmall, midX, midY+4, ColorTextPrimary, text.AlignCenter)
	}
}

// drawParticles renders particles
func (r *Renderer) drawParticles(screen *ebiten.Image, ps *ParticleSystem) {
	for _, p := range ps.particles {
		if p.Alpha <= 0.01 {
			continue
		}

		c := color.RGBA{
			R: p.Color.R,
			G: p.Color.G,
			B: p.Color.B,
			A: uint8(float64(p.Color.A) * p.Alpha),
		}

		switch p.Kind {
		case ParticleKindAmbient:
			vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size), c, true)
		case ParticleKindCurveTrail:
			vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size), c, true)
		case ParticleKindAttention:
			// Attention packets have a small halo
			halo := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: uint8(float64(c.A) * 0.4)}
			vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size*2.0), halo, true)
			vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size), c, true)
		}
	}
}

// drawHeader draws the top cinematic title bar
func (r *Renderer) drawHeader(screen *ebiten.Image, anim *AnimationManager) {
	// Top banner background
	vector.FillRect(screen, 0, 0, ScreenWidth, HeaderHeight, ColorBgPanelDark, true)
	vector.StrokeLine(screen, 0, HeaderHeight, ScreenWidth, HeaderHeight, 1.0, ColorBorder, false)

	// Left: Title and subtitle
	r.DrawText(screen, "ATTENTION", r.faceHeading, 36, 30, ColorTextPrimary)
	r.DrawText(screen, "The Mathematics of Changing Context", r.faceSmall, 160, 30, ColorTextSecondary)

	// Center-left: Current scene indicator
	meta := SceneMetas[anim.CurrentScene]
	sceneBadge := fmt.Sprintf("[%s / 05]", meta.Number)
	r.DrawText(screen, sceneBadge, r.faceMono, 560, 30, ColorAccent)
	r.DrawText(screen, meta.Title, r.faceHeading, 700, 30, ColorTextPrimary)

	// Mathematical explanation of current scene
	r.DrawText(screen, meta.Concept, r.faceSmall, 36, 56, ColorTextMuted)

	// Right: Quick key hints
	r.DrawText(screen, "Press [H] Help  |  [M] HUD  |  [C] Causal Mask  |  [Space] Pause", r.faceMonoSmall, ScreenWidth-540, 38, ColorTextSecondary)
}

// drawHUD renders the comprehensive mathematical inspector on the right
func (r *Renderer) drawHUD(screen *ebiten.Image, sim *Simulation, anim *AnimationManager) {
	x := float32(ScreenWidth - HUDWidth - HUDMargin)
	y := float32(HeaderHeight + 20)
	w := float32(HUDWidth)
	h := float32(ScreenHeight - HeaderHeight - FooterHeight - 40)

	// Dark glass container
	vector.FillRect(screen, x, y, w, h, ColorBgPanel, true)
	vector.StrokeRect(screen, x, y, w, h, 1.2, ColorBorder, false)

	padX := float64(x + 22)
	currY := float64(y + 24)

	// Header
	r.DrawText(screen, "ATTENTION TENSOR INSPECTOR", r.faceHeading, padX, currY, ColorAccent)
	currY += 22
	r.DrawText(screen, "EXP-019 / Scaled Dot-Product Engine", r.faceSmall, padX, currY, ColorTextMuted)
	currY += 28

	// Multi-head specification (PDF Section 03)
	r.DrawText(screen, "MODEL HYPERPARAMETERS", r.faceMonoSmall, padX, currY, ColorTextSecondary)
	currY += 16
	r.DrawText(screen, sim.AttentionEngine.MultiHeadSummary(), r.faceMonoSmall, padX, currY, ColorTextMuted)
	currY += 26

	// Causal Mask Indicator (PDF Section 02)
	causalStatus := "CAUSAL MASK: OFF (Bidirectional)"
	causalCol := ColorTextSecondary
	if anim.CausalMask {
		causalStatus = "CAUSAL MASK: ACTIVE (Lower Triangular M)"
		causalCol = ColorAccentGreen
	}
	vector.FillRect(screen, float32(padX-4), float32(currY-12), float32(w-36), 22, ColorBgPanelDark, true)
	vector.StrokeRect(screen, float32(padX-4), float32(currY-12), float32(w-36), 22, 1.0, causalCol, false)
	r.DrawText(screen, causalStatus, r.faceMonoSmall, padX+4, currY+2, causalCol)
	currY += 34

	// Separator
	vector.StrokeLine(screen, float32(padX), float32(currY), float32(x+w-22), float32(currY), 1.0, ColorBorder, false)
	currY += 16

	// Q and K Vectors
	r.DrawText(screen, "EMBEDDING MATRICES (Q, K, V)", r.faceMonoSmall, padX, currY, ColorTextSecondary)
	currY += 18
	r.DrawText(screen, "Token A : Q=[1, 0]  K=[1, 0]  V=[1, 0]", r.faceMonoSmall, padX, currY, ColorTokenA)
	currY += 16
	r.DrawText(screen, "Token B : Q=[0, 1]  K=[0, 1]  V=[0, 2]", r.faceMonoSmall, padX, currY, ColorTokenB)
	currY += 16
	if anim.CurrentScene >= Scene03NewContext {
		r.DrawText(screen, "Token C : Q=[1, 1]  K=[1, 1]  V=[2, 1]", r.faceMonoSmall, padX, currY, ColorTokenC)
	} else {
		r.DrawText(screen, "Token C : [Not yet entered into sequence]", r.faceMonoSmall, padX, currY, ColorTextMuted)
	}
	currY += 28

	// Scores Matrix S = QK^T / sqrt(d_k)
	r.DrawText(screen, "1. SCALED SCORES:  S = Q K^T / √d_k", r.faceMonoSmall, padX, currY, ColorAccent)
	currY += 18
	r.drawMatrixBlock(screen, padX, currY, sim.CurrentResult.Scores, []string{"A", "B", "C"}, false)
	currY += float64(len(sim.CurrentResult.Scores)*18 + 14)

	// Softmax Attention Weights Matrix A = softmax(S)
	r.DrawText(screen, "2. ATTENTION WEIGHTS:  A = softmax(S)", r.faceMonoSmall, padX, currY, ColorAccent)
	currY += 18
	r.drawMatrixBlock(screen, padX, currY, sim.CurrentResult.Weights, []string{"A", "B", "C"}, true)
	currY += float64(len(sim.CurrentResult.Weights)*18 + 14)

	// Row Sums Verification
	r.DrawText(screen, "VERIFICATION:  Σ_j  A_ij  ==  1.000", r.faceMonoSmall, padX, currY, ColorTextSecondary)
	currY += 16
	sumStr := ""
	for i, s := range sim.CurrentResult.RowSums {
		name := string('A' + rune(i))
		sumStr += fmt.Sprintf("Row %s: %.4f   ", name, s)
	}
	r.DrawText(screen, sumStr, r.faceMonoSmall, padX, currY, ColorAccentGreen)
	currY += 26

	// Output Contextual Vectors O = A * V
	r.DrawText(screen, "3. CONTEXTUAL OUTPUTS:  O = A * V", r.faceMonoSmall, padX, currY, ColorAccent)
	currY += 18
	for i := 0; i < len(sim.CurrentResult.Output); i++ {
		out := sim.CurrentResult.Output[i]
		tokenName := string('A' + rune(i))
		var col color.RGBA
		switch i {
		case 0:
			col = ColorTokenA
		case 1:
			col = ColorTokenB
		case 2:
			col = ColorTokenC
		}
		str := fmt.Sprintf("O_%s = [%.3f,  %.3f]", tokenName, out[0], out[1])
		r.DrawText(screen, str, r.faceMonoSmall, padX, currY, col)
		currY += 16
	}
	currY += 20

	// 2D Representation Subspace Inset Plot
	r.DrawText(screen, "2D EMBEDDING SUBSPACE PROJECTION", r.faceMonoSmall, padX, currY, ColorTextSecondary)
	currY += 16
	r.drawSubspacePlot(screen, padX+10, currY+10, sim)
}

// drawMatrixBlock renders a formatted numerical matrix
func (r *Renderer) drawMatrixBlock(screen *ebiten.Image, startX, startY float64, mat [][]float64, labels []string, isWeight bool) {
	n := len(mat)
	cellW := 80.0
	cellH := 18.0

	// Background box
	vector.FillRect(screen, float32(startX), float32(startY-4), float32(float64(n)*cellW+30), float32(float64(n)*cellH+12), ColorBgPanelDark, true)

	for i := 0; i < n; i++ {
		// Row label
		rowLabel := fmt.Sprintf("[%s]", labels[i])
		r.DrawText(screen, rowLabel, r.faceMonoSmall, startX+4, startY+float64(i)*cellH+10, ColorTextMuted)

		for j := 0; j < len(mat[i]); j++ {
			val := mat[i][j]
			var valStr string
			if val <= -1e8 {
				valStr = "  -inf"
			} else {
				valStr = fmt.Sprintf("%7.3f", val)
			}

			// Highlight diagonal vs cross attention
			col := ColorTextPrimary
			if isWeight {
				if i == j {
					col = ColorTextSecondary
				} else if val > 0.25 {
					col = ColorAccent
				}
			}

			r.DrawText(screen, valStr, r.faceMonoSmall, startX+32+float64(j)*cellW, startY+float64(i)*cellH+10, col)
		}
	}
}

// drawSubspacePlot draws an inset 2D vector coordinate plot
func (r *Renderer) drawSubspacePlot(screen *ebiten.Image, cx, cy float64, sim *Simulation) {
	plotW := float32(HUDWidth - 75)
	plotH := float32(110)

	vector.FillRect(screen, float32(cx), float32(cy), plotW, plotH, ColorBgPanelDark, true)
	vector.StrokeRect(screen, float32(cx), float32(cy), plotW, plotH, 1.0, ColorBorder, false)

	// Center axes
	originX := float32(cx + 40)
	originY := float32(cy + float64(plotH) - 20)
	vector.StrokeLine(screen, originX, float32(cy), originX, float32(cy+float64(plotH)), 1.0, ColorGrid, false)
	vector.StrokeLine(screen, float32(cx), originY, float32(cx+float64(plotW)), originY, 1.0, ColorGrid, false)

	scale := float32(32.0)

	// Plot Initial V vectors (circles) and Contextual O vectors (diamonds/squares)
	for i := 0; i < len(sim.AttentionEngine.BaseV); i++ {
		orig := sim.AttentionEngine.BaseV[i]
		ox := originX + float32(orig[0])*scale
		oy := originY - float32(orig[1])*scale

		var col color.RGBA
		switch i {
		case 0:
			col = ColorTokenA
		case 1:
			col = ColorTokenB
		case 2:
			col = ColorTokenC
		}

		// Initial V: Hollow ring
		vector.StrokeCircle(screen, ox, oy, 4.0, 1.2, color.RGBA{R: col.R, G: col.G, B: col.B, A: 120}, true)

		// Contextual Output O: Solid filled circle with connecting trajectory line
		if i < len(sim.CurrentResult.Output) {
			out := sim.CurrentResult.Output[i]
			outX := originX + float32(out[0])*scale
			outY := originY - float32(out[1])*scale

			// Trajectory line from V -> O
			vector.StrokeLine(screen, ox, oy, outX, outY, 1.0, color.RGBA{R: col.R, G: col.G, B: col.B, A: 160}, true)
			vector.FillCircle(screen, outX, outY, 4.5, col, true)

			lbl := fmt.Sprintf("O_%s", string('A'+rune(i)))
			r.DrawText(screen, lbl, r.faceMonoSmall, float64(outX+6), float64(outY+3), col)
		}
	}
}

// drawAftermathOverlay draws the climactic quote and title in Scene 05
func (r *Renderer) drawAftermathOverlay(screen *ebiten.Image, anim *AnimationManager) {
	prog := anim.SceneProgress()
	alpha := EaseInOutQuad(prog)

	textCol := color.RGBA{
		R: ColorTextPrimary.R,
		G: ColorTextPrimary.G,
		B: ColorTextPrimary.B,
		A: uint8(float64(ColorTextPrimary.A) * alpha),
	}

	accentCol := color.RGBA{
		R: ColorAccent.R,
		G: ColorAccent.G,
		B: ColorAccent.B,
		A: uint8(float64(ColorAccent.A) * alpha),
	}

	midX := float64((ScreenWidth - HUDWidth) / 2)
	midY := float64(ScreenHeight * 0.42)

	// Epilogue quote
	r.DrawTextAligned(screen, "\"The same tokens. A different context.\"", r.faceHero, midX, midY, textCol, text.AlignCenter)

	// Subtitle
	r.DrawTextAligned(screen, "ATTENTION MECHANISM & TRANSFORMER", r.faceHeading, midX, midY+50, accentCol, text.AlignCenter)
	r.DrawTextAligned(screen, "EXP-019 / Scaled Dot-Product & Contextual Geometry", r.faceSmall, midX, midY+85, ColorTextMuted, text.AlignCenter)
}

// drawFooter draws the timeline scrubber, playback controls, and scene indicators
func (r *Renderer) drawFooter(screen *ebiten.Image, anim *AnimationManager) {
	y := float32(ScreenHeight - FooterHeight)
	w := float32(ScreenWidth)
	h := float32(FooterHeight)

	// Footer background
	vector.FillRect(screen, 0, y, w, h, ColorBgPanelDark, true)
	vector.StrokeLine(screen, 0, y, w, y, 1.0, ColorBorder, false)

	// Timeline scrubber across the bottom
	barX := float32(40)
	barY := float32(y + 14)
	barW := float32(ScreenWidth - 80)
	barH := float32(6)

	// Background track
	vector.FillRect(screen, barX, barY, barW, barH, ColorGrid, true)

	// Segments for 5 scenes
	segmentW := barW / 5.0
	for i := 1; i <= 5; i++ {
		segX := barX + float32(i-1)*segmentW
		isActive := int(anim.CurrentScene) == i
		isPast := int(anim.CurrentScene) > i

		if isPast {
			vector.FillRect(screen, segX, barY, segmentW-2, barH, ColorAccent, true)
		} else if isActive {
			fillProg := float32(anim.SceneProgress())
			vector.FillRect(screen, segX, barY, (segmentW-2)*fillProg, barH, ColorTokenA, true)
		}

		// Divider notch
		vector.StrokeLine(screen, segX, barY-3, segX, barY+barH+3, 1.0, ColorBorder, false)

		// Scene label
		lbl := fmt.Sprintf("0%d", i)
		lblCol := ColorTextMuted
		if isActive {
			lblCol = ColorTextPrimary
		}
		r.DrawText(screen, lbl, r.faceMonoSmall, float64(segX+8), float64(barY+22), lblCol)
	}

	// Play / Pause status badge
	statusStr := "[PLAYING]"
	if anim.IsPaused {
		statusStr = "[PAUSED]"
	}
	r.DrawText(screen, statusStr, r.faceMono, float64(barX+240), float64(y+38), ColorAccent)

	// Speed badge
	speedStr := fmt.Sprintf("Speed: %.1fx", anim.PlaybackSpeed)
	r.DrawText(screen, speedStr, r.faceMono, float64(barX+360), float64(y+38), ColorTextSecondary)

	// Navigation tips
	r.DrawText(screen, "Controls: [Space] Play/Pause  |  [←/→] Prev/Next Scene  |  [1-5] Jump  |  [R] Restart  |  [↑/↓] Speed", r.faceSmall, float64(barX+520), float64(y+38), ColorTextMuted)
}

// drawHelpOverlay draws the full keybinding help card
func (r *Renderer) drawHelpOverlay(screen *ebiten.Image) {
	cardW := float32(640)
	cardH := float32(460)
	cx := float32(ScreenWidth-cardW) / 2
	cy := float32(ScreenHeight-cardH) / 2

	// Modal shadow and background
	vector.FillRect(screen, cx-4, cy-4, cardW+8, cardH+8, color.RGBA{R: 0, G: 0, B: 0, A: 220}, true)
	vector.FillRect(screen, cx, cy, cardW, cardH, ColorBgPanel, true)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1.5, ColorAccent, false)

	padX := float64(cx + 36)
	currY := float64(cy + 40)

	r.DrawText(screen, "ATTENTION VISUALIZATION CONTROLS", r.faceHeading, padX, currY, ColorAccent)
	currY += 26
	r.DrawText(screen, "Interactive scientific animation controls and keybindings", r.faceSmall, padX, currY, ColorTextSecondary)
	currY += 36

	shortcuts := []struct {
		Key  string
		Desc string
	}{
		{"SPACE", "Toggle Play / Pause animation"},
		{"R", "Restart animation from Scene 01"},
		{"LEFT / RIGHT", "Step to Previous / Next Scene"},
		{"1, 2, 3, 4, 5", "Jump directly to Scene 01 - 05"},
		{"UP / DOWN or [ / ]", "Increase / Decrease animation playback speed (0.25x - 3.0x)"},
		{"M", "Toggle Attention Tensor Inspector (Mathematical HUD)"},
		{"C", "Toggle Causal Masking (M lower-triangular matrix, PDF Sec. 02)"},
		{"H", "Toggle this Help & Keybindings Overlay"},
		{"ESC", "Close overlays"},
	}

	for _, s := range shortcuts {
		vector.FillRect(screen, float32(padX), float32(currY-12), 150, 22, ColorBgPanelDark, true)
		vector.StrokeRect(screen, float32(padX), float32(currY-12), 150, 22, 1.0, ColorBorder, false)
		r.DrawText(screen, s.Key, r.faceMonoSmall, padX+10, currY+2, ColorTextPrimary)
		r.DrawText(screen, s.Desc, r.faceBody, padX+170, currY+4, ColorTextSecondary)
		currY += 30
	}

	currY += 14
	r.DrawTextAligned(screen, "Press [H] or [ESC] to resume", r.faceSmall, float64(cx+cardW/2), currY, ColorAccentGreen, text.AlignCenter)
}

