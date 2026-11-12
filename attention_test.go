package main

import (
	"math"
	"testing"
)

const floatTol = 1e-3

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestScaledDotProductAttention_3Tokens(t *testing.T) {
	engine := NewAttentionEngine()
	res := engine.ComputeScaledDotProduct(3, false)

	// Verify Dot Product matrix QK^T:
	// Q = [[1,0], [0,1], [1,1]], K = [[1,0], [0,1], [1,1]]
	// QK^T = [[1, 0, 1], [0, 1, 1], [1, 1, 2]]
	expectedDot := [][]float64{
		{1.0, 0.0, 1.0},
		{0.0, 1.0, 1.0},
		{1.0, 1.0, 2.0},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if !almostEqual(res.DotProducts[i][j], expectedDot[i][j], floatTol) {
				t.Errorf("DotProduct[%d][%d] = %f, expected %f", i, j, res.DotProducts[i][j], expectedDot[i][j])
			}
		}
	}

	// Scaled scores: S = QK^T / sqrt(2)
	sqrt2 := math.Sqrt(2.0)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			expectedScore := expectedDot[i][j] / sqrt2
			if !almostEqual(res.Scores[i][j], expectedScore, floatTol) {
				t.Errorf("Scores[%d][%d] = %f, expected %f", i, j, res.Scores[i][j], expectedScore)
			}
		}
	}

	// Softmax Weights:
	// Row 0: [0.4011, 0.1978, 0.4011]
	// Row 1: [0.1978, 0.4011, 0.4011]
	// Row 2: [0.2483, 0.2483, 0.5035]
	expectedWeights := [][]float64{
		{0.4011, 0.1978, 0.4011},
		{0.1978, 0.4011, 0.4011},
		{0.2483, 0.2483, 0.5035},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if !almostEqual(res.Weights[i][j], expectedWeights[i][j], 0.005) {
				t.Errorf("Weight[%d][%d] = %f, expected %f", i, j, res.Weights[i][j], expectedWeights[i][j])
			}
		}
		// Row sum must be 1.0
		if !almostEqual(res.RowSums[i], 1.0, 1e-5) {
			t.Errorf("RowSum[%d] = %f, expected 1.0", i, res.RowSums[i])
		}
	}

	// Contextual Output O = A * V:
	// V = [[1, 0], [0, 2], [2, 1]]
	// O_0 = 0.4011*[1,0] + 0.1978*[0,2] + 0.4011*[2,1] = [1.2033, 0.7967]
	// O_1 = 0.1978*[1,0] + 0.4011*[0,2] + 0.4011*[2,1] = [1.0000, 1.2033]
	// O_2 = 0.2483*[1,0] + 0.2483*[0,2] + 0.5035*[2,1] = [1.2553, 1.0001]
	expectedOut := [][]float64{
		{1.2033, 0.7967},
		{1.0000, 1.2033},
		{1.2553, 1.0001},
	}

	for i := 0; i < 3; i++ {
		for d := 0; d < 2; d++ {
			if !almostEqual(res.Output[i][d], expectedOut[i][d], 0.01) {
				t.Errorf("Output[%d][%d] = %f, expected %f", i, d, res.Output[i][d], expectedOut[i][d])
			}
		}
	}
}

func TestCausalMasking(t *testing.T) {
	engine := NewAttentionEngine()
	res := engine.ComputeScaledDotProduct(3, true)

	// Token 0 can only attend to itself
	if !almostEqual(res.Weights[0][0], 1.0, floatTol) ||
		!almostEqual(res.Weights[0][1], 0.0, floatTol) ||
		!almostEqual(res.Weights[0][2], 0.0, floatTol) {
		t.Errorf("Causal mask row 0 failed: %v", res.Weights[0])
	}

	// Token 1 can attend to 0 and 1, but not 2
	if !almostEqual(res.Weights[1][2], 0.0, floatTol) {
		t.Errorf("Causal mask row 1 attended to future token: %f", res.Weights[1][2])
	}
	if !almostEqual(res.Weights[1][0], 0.3302, 0.01) || !almostEqual(res.Weights[1][1], 0.6698, 0.01) {
		t.Errorf("Causal mask row 1 values incorrect: %v", res.Weights[1])
	}

	// Token 2 can attend to all
	if !almostEqual(res.Weights[2][0], 0.2483, 0.01) ||
		!almostEqual(res.Weights[2][1], 0.2483, 0.01) ||
		!almostEqual(res.Weights[2][2], 0.5035, 0.01) {
		t.Errorf("Causal mask row 2 values incorrect: %v", res.Weights[2])
	}
}

func TestAttentionRedistributionEffect(t *testing.T) {
	engine := NewAttentionEngine()

	// Initial 2-token state (Tokens A and B)
	res2 := engine.ComputeScaledDotProduct(2, false)
	weightAB_initial := res2.Weights[0][1]

	// 3-token state (Token C added)
	res3 := engine.ComputeScaledDotProduct(3, false)
	weightAB_afterC := res3.Weights[0][1]
	weightAC := res3.Weights[0][2]

	t.Logf("Initial alpha_AB: %.4f", weightAB_initial)
	t.Logf("After Token C alpha_AB: %.4f, alpha_AC: %.4f", weightAB_afterC, weightAC)

	// Attention between A and B must significantly drop when C arrives
	if weightAB_afterC >= weightAB_initial {
		t.Errorf("Expected alpha_AB to decrease after Token C enters: initial=%f, after=%f", weightAB_initial, weightAB_afterC)
	}

	// Attention to C must exceed attention between A and B
	if weightAC <= weightAB_afterC {
		t.Errorf("Expected Token C to capture dominant attention: alpha_AC=%f, alpha_AB=%f", weightAC, weightAB_afterC)
	}
}

func TestSimulationLoopAllScenes(t *testing.T) {
	sim := NewSimulation()
	anim := NewAnimationManager()

	// Step through all 5 scenes and simulate multiple frames in each scene
	for scene := Scene01InitialDistance; scene <= Scene05Aftermath; scene++ {
		anim.SetScene(scene)
		for step := 0; step < 120; step++ {
			dt := 1.0 / 60.0
			anim.Update(dt)
			sim.Update(dt, anim)
		}
	}

	// Test with Causal Mask active
	anim.CausalMask = true
	for scene := Scene01InitialDistance; scene <= Scene05Aftermath; scene++ {
		anim.SetScene(scene)
		for step := 0; step < 60; step++ {
			dt := 1.0 / 60.0
			anim.Update(dt)
			sim.Update(dt, anim)
		}
	}
}

