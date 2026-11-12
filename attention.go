package main

import (
	"fmt"
	"math"
)

// AttentionEngine encapsulates scaled dot-product attention calculations
// directly derived from the engineering note: Attention_Mechanism_Transformer_Materi.pdf
type AttentionEngine struct {
	Dk     float64   // Key dimension (2.0)
	Dv     float64   // Value dimension (2.0)
	Dmodel int       // Model dimension (4)
	Heads  int       // Number of attention heads (2)

	// Standard reference matrices from EXP-019 note:
	// Q = [[1, 0], [0, 1], [1, 1]]
	// K = [[1, 0], [0, 1], [1, 1]]
	// V = [[1, 0], [0, 2], [2, 1]]
	BaseQ [][]float64
	BaseK [][]float64
	BaseV [][]float64
}

// NewAttentionEngine instantiates the engine with parameters matching the PDF note
func NewAttentionEngine() *AttentionEngine {
	return &AttentionEngine{
		Dk:     2.0,
		Dv:     2.0,
		Dmodel: 4,
		Heads:  2,
		BaseQ: [][]float64{
			{1.0, 0.0}, // Token A (Row 0)
			{0.0, 1.0}, // Token B (Row 1)
			{1.0, 1.0}, // Token C (Row 2)
		},
		BaseK: [][]float64{
			{1.0, 0.0},
			{0.0, 1.0},
			{1.0, 1.0},
		},
		BaseV: [][]float64{
			{1.0, 0.0},
			{0.0, 2.0},
			{2.0, 1.0},
		},
	}
}

// AttentionResult packages all intermediate and final mathematical tensors
type AttentionResult struct {
	NumTokens    int
	Q            [][]float64 // Query matrix (n x d_k)
	K            [][]float64 // Key matrix (n x d_k)
	V            [][]float64 // Value matrix (n x d_v)
	DotProducts  [][]float64 // QK^T
	Scores       [][]float64 // S = QK^T / sqrt(d_k)
	MaskedScores [][]float64 // S_masked with causal mask applied if active
	Weights      [][]float64 // A = softmax(Scores)
	Output       [][]float64 // O = A * V (Contextual representations)
	RowSums      []float64   // Verification that sum_j A_ij == 1.0
	IsCausal     bool
}

// ComputeScaledDotProduct computes full scaled dot-product attention
// for n tokens (2 or 3), with optional causal mask M and optional alpha blending.
func (ae *AttentionEngine) ComputeScaledDotProduct(numTokens int, causalMask bool) AttentionResult {
	if numTokens < 1 {
		numTokens = 1
	}
	if numTokens > 3 {
		numTokens = 3
	}

	res := AttentionResult{
		NumTokens:    numTokens,
		Q:            make([][]float64, numTokens),
		K:            make([][]float64, numTokens),
		V:            make([][]float64, numTokens),
		DotProducts:  make([][]float64, numTokens),
		Scores:       make([][]float64, numTokens),
		MaskedScores: make([][]float64, numTokens),
		Weights:      make([][]float64, numTokens),
		Output:       make([][]float64, numTokens),
		RowSums:      make([]float64, numTokens),
		IsCausal:     causalMask,
	}

	// Copy submatrices
	for i := 0; i < numTokens; i++ {
		res.Q[i] = append([]float64(nil), ae.BaseQ[i]...)
		res.K[i] = append([]float64(nil), ae.BaseK[i]...)
		res.V[i] = append([]float64(nil), ae.BaseV[i]...)

		res.DotProducts[i] = make([]float64, numTokens)
		res.Scores[i] = make([]float64, numTokens)
		res.MaskedScores[i] = make([]float64, numTokens)
		res.Weights[i] = make([]float64, numTokens)
		res.Output[i] = make([]float64, int(ae.Dv))
	}

	sqrtDk := math.Sqrt(ae.Dk)

	// 1. Compute Q * K^T and Scores S = QK^T / sqrt(d_k)
	for i := 0; i < numTokens; i++ {
		for j := 0; j < numTokens; j++ {
			dot := 0.0
			for d := 0; d < int(ae.Dk); d++ {
				dot += res.Q[i][d] * res.K[j][d]
			}
			res.DotProducts[i][j] = dot
			score := dot / sqrtDk
			res.Scores[i][j] = score

			// Check causal mask: token i cannot attend to future token j > i
			if causalMask && j > i {
				res.MaskedScores[i][j] = -1e9 // Effective -infinity
			} else {
				res.MaskedScores[i][j] = score
			}
		}
	}

	// 2. Compute Softmax row by row: softmax(z)_j = exp(z_j - max) / sum(exp(z - max))
	for i := 0; i < numTokens; i++ {
		scoresToSoftmax := res.Scores[i]
		if causalMask {
			scoresToSoftmax = res.MaskedScores[i]
		}

		// Find max for numerical stability
		maxVal := scoresToSoftmax[0]
		for j := 1; j < numTokens; j++ {
			if scoresToSoftmax[j] > maxVal {
				maxVal = scoresToSoftmax[j]
			}
		}

		sumExp := 0.0
		expVals := make([]float64, numTokens)
		for j := 0; j < numTokens; j++ {
			expVals[j] = math.Exp(scoresToSoftmax[j] - maxVal)
			sumExp += expVals[j]
		}

		rowSum := 0.0
		for j := 0; j < numTokens; j++ {
			res.Weights[i][j] = expVals[j] / sumExp
			rowSum += res.Weights[i][j]
		}
		res.RowSums[i] = rowSum
	}

	// 3. Compute Output = Weights * V
	for i := 0; i < numTokens; i++ {
		for d := 0; d < int(ae.Dv); d++ {
			val := 0.0
			for j := 0; j < numTokens; j++ {
				val += res.Weights[i][j] * res.V[j][d]
			}
			res.Output[i][d] = val
		}
	}

	return res
}

// InterpolatedAttention calculates smoothly blended attention as Token C introduces its presence [0.0 -> 1.0]
// This allows continuous, mathematically sound transitions during Scene 03 & 04.
// Always returns a 3x3 result so all 3 tokens have defined attention weights throughout.
func (ae *AttentionEngine) InterpolatedAttention(tokenCPresence float64, causalMask bool) AttentionResult {
	// Dynamic Q, K, V where Token C's vector magnitudes scale smoothly with presence
	p := math.Max(0.0, math.Min(1.0, tokenCPresence))
	numTokens := 3

	res := AttentionResult{
		NumTokens:    numTokens,
		Q:            make([][]float64, numTokens),
		K:            make([][]float64, numTokens),
		V:            make([][]float64, numTokens),
		DotProducts:  make([][]float64, numTokens),
		Scores:       make([][]float64, numTokens),
		MaskedScores: make([][]float64, numTokens),
		Weights:      make([][]float64, numTokens),
		Output:       make([][]float64, numTokens),
		RowSums:      make([]float64, numTokens),
		IsCausal:     causalMask,
	}

	for i := 0; i < numTokens; i++ {
		res.Q[i] = make([]float64, int(ae.Dk))
		res.K[i] = make([]float64, int(ae.Dk))
		res.V[i] = make([]float64, int(ae.Dv))

		scale := 1.0
		if i == 2 {
			scale = p
		}
		for d := 0; d < int(ae.Dk); d++ {
			res.Q[i][d] = ae.BaseQ[i][d] * scale
			res.K[i][d] = ae.BaseK[i][d] * scale
		}
		for d := 0; d < int(ae.Dv); d++ {
			res.V[i][d] = ae.BaseV[i][d] * scale
		}

		res.DotProducts[i] = make([]float64, numTokens)
		res.Scores[i] = make([]float64, numTokens)
		res.MaskedScores[i] = make([]float64, numTokens)
		res.Weights[i] = make([]float64, numTokens)
		res.Output[i] = make([]float64, int(ae.Dv))
	}

	sqrtDk := math.Sqrt(ae.Dk)

	// Compute dot products and scores
	for i := 0; i < numTokens; i++ {
		for j := 0; j < numTokens; j++ {
			dot := 0.0
			for d := 0; d < int(ae.Dk); d++ {
				dot += res.Q[i][d] * res.K[j][d]
			}
			res.DotProducts[i][j] = dot
			score := dot / sqrtDk

			// When token C is partial, suppress its cross-attention scores accordingly
			if (i == 2 || j == 2) && p < 1.0 {
				// Suppress by log(p) so that softmax naturally vanishes as p -> 0
				if p > 0.0001 {
					score += math.Log(p) * 2.0
				} else {
					score = -1e9
				}
			}

			res.Scores[i][j] = score
			if causalMask && j > i {
				res.MaskedScores[i][j] = -1e9
			} else {
				res.MaskedScores[i][j] = score
			}
		}
	}

	// Softmax
	for i := 0; i < numTokens; i++ {
		scoresToSoftmax := res.Scores[i]
		if causalMask {
			scoresToSoftmax = res.MaskedScores[i]
		}

		maxVal := scoresToSoftmax[0]
		for j := 1; j < numTokens; j++ {
			if scoresToSoftmax[j] > maxVal {
				maxVal = scoresToSoftmax[j]
			}
		}

		sumExp := 0.0
		expVals := make([]float64, numTokens)
		for j := 0; j < numTokens; j++ {
			expVals[j] = math.Exp(scoresToSoftmax[j] - maxVal)
			sumExp += expVals[j]
		}

		rowSum := 0.0
		for j := 0; j < numTokens; j++ {
			res.Weights[i][j] = expVals[j] / sumExp
			rowSum += res.Weights[i][j]
		}
		res.RowSums[i] = rowSum
	}

	// Output = Weights * V
	for i := 0; i < numTokens; i++ {
		for d := 0; d < int(ae.Dv); d++ {
			val := 0.0
			for j := 0; j < numTokens; j++ {
				val += res.Weights[i][j] * res.V[j][d]
			}
			res.Output[i][d] = val
		}
	}

	return res
}

// MultiHeadSummary formats multi-head parameters specified in Section 03 of the PDF
func (ae *AttentionEngine) MultiHeadSummary() string {
	return fmt.Sprintf("Multi-Head Config: d_model=%d, h=%d, d_k=%d, d_v=%d | W_O in R^(%dx%d)",
		ae.Dmodel, ae.Heads, int(ae.Dk), int(ae.Dv), ae.Heads*int(ae.Dv), ae.Dmodel)
}

