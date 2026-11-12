package main

import (
	"math"
)

// Easing utilities for cinematic, mathematically smooth motion
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func Clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func Smoothstep(edge0, edge1, x float64) float64 {
	t := Clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

func EaseInOutCubic(t float64) float64 {
	t = Clamp(t, 0.0, 1.0)
	if t < 0.5 {
		return 4.0 * t * t * t
	}
	p := 2.0*t - 2.0
	return 0.5*p*p*p + 1.0
}

func EaseOutQuad(t float64) float64 {
	t = Clamp(t, 0.0, 1.0)
	return 1.0 - (1.0-t)*(1.0-t)
}

func EaseInOutQuad(t float64) float64 {
	t = Clamp(t, 0.0, 1.0)
	if t < 0.5 {
		return 2.0 * t * t
	}
	return -1.0 + (4.0-2.0*t)*t
}

// AnimationManager coordinates timeline, scene transitions, and playback control
type AnimationManager struct {
	CurrentScene  SceneID
	SceneTime     float64
	TotalTime     float64
	PlaybackSpeed float64
	IsPaused      bool
	ShowHUD       bool
	CausalMask    bool
	ShowHelp      bool

	// Scene title transition opacity
	TitleAlpha float64
}

// NewAnimationManager creates default sequencer
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		CurrentScene:  Scene01InitialDistance,
		SceneTime:     0.0,
		TotalTime:     0.0,
		PlaybackSpeed: 1.0,
		IsPaused:      false,
		ShowHUD:       true,
		CausalMask:    false,
		ShowHelp:      false,
		TitleAlpha:    1.0,
	}
}

// SceneProgress returns normalized progress [0.0 - 1.0] of current scene
func (am *AnimationManager) SceneProgress() float64 {
	dur := SceneDurations[am.CurrentScene]
	if dur <= 0 {
		return 1.0
	}
	return Clamp(am.SceneTime/dur, 0.0, 1.0)
}

// Update advances the timeline
func (am *AnimationManager) Update(rawDT float64) {
	if am.IsPaused {
		return
	}

	dt := rawDT * am.PlaybackSpeed
	am.SceneTime += dt
	am.TotalTime += dt

	dur := SceneDurations[am.CurrentScene]
	prog := am.SceneProgress()

	// Title alpha calculation: smooth fade in during first 20%, stay visible, subtle fade at end
	if prog < 0.20 {
		am.TitleAlpha = Smoothstep(0.0, 0.20, prog)
	} else if prog > 0.88 {
		am.TitleAlpha = 1.0 - Smoothstep(0.88, 1.0, prog)
	} else {
		am.TitleAlpha = 1.0
	}

	// Automatic progression to next scene when duration reached
	if am.SceneTime >= dur {
		if am.CurrentScene < Scene05Aftermath {
			am.SetScene(am.CurrentScene + 1)
		} else {
			// At Scene 5, loop back to Scene 1 smoothly after staying in aftermath
			am.SetScene(Scene01InitialDistance)
		}
	}
}

// SetScene transitions directly to a designated scene
func (am *AnimationManager) SetScene(scene SceneID) {
	if scene < Scene01InitialDistance {
		scene = Scene01InitialDistance
	}
	if scene > Scene05Aftermath {
		scene = Scene05Aftermath
	}
	am.CurrentScene = scene
	am.SceneTime = 0.0
	am.TitleAlpha = 0.0
}

// NextScene advances to the next scene
func (am *AnimationManager) NextScene() {
	if am.CurrentScene < Scene05Aftermath {
		am.SetScene(am.CurrentScene + 1)
	}
}

// PrevScene steps back to the previous scene
func (am *AnimationManager) PrevScene() {
	if am.CurrentScene > Scene01InitialDistance {
		am.SetScene(am.CurrentScene - 1)
	} else {
		am.SceneTime = 0.0
	}
}

// TogglePause toggles playback state
func (am *AnimationManager) TogglePause() {
	am.IsPaused = !am.IsPaused
}

// Restart restarts the full animation from Scene 01
func (am *AnimationManager) Restart() {
	am.SetScene(Scene01InitialDistance)
	am.TotalTime = 0.0
}

// AdjustSpeed changes the playback rate
func (am *AnimationManager) AdjustSpeed(delta float64) {
	newSpeed := am.PlaybackSpeed + delta
	if newSpeed < 0.25 {
		newSpeed = 0.25
	}
	if newSpeed > 3.0 {
		newSpeed = 3.0
	}
	am.PlaybackSpeed = math.Round(newSpeed*10.0) / 10.0
}

