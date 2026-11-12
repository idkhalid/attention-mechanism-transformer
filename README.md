# Attention: The Mathematics of Changing Context

An interactive 2D visualization built in Go (using Ebitengine) that visualizes **Scaled Dot-Product Attention** in Transformer architectures.

The project translates the linear algebra behind self-attention into wave mechanics: showing how two isolated token representations interact, and how adding a new token to the sequence redistributes attention weights and alters their contextual representations.

The mathematical computations follow the reference notes in `Attention-Mechanism-Transformer.pdf` (EXP-019).

---

## Visual Concept

The simulation runs through 5 sequential scenes:

1. **Scene 01 — Initial Distance**: Two isolated tokens (Token A and Token B) enter from opposite sides as independent waveforms with no cross-token interaction ($\alpha_{AB} = 0$).
2. **Scene 02 — Convergence**: Token A and Token B move toward the center line. Pairwise attention forms ($\alpha_{AA} \approx 0.67, \alpha_{AB} \approx 0.33$). Their frequencies begin to phase-lock and glowing energy filaments connect the two curves.
3. **Scene 03 — A New Context**: Token C ($Q=[1, 1], K=[1, 1], V=[2, 1]$) enters the sequence unexpectedly. The score matrix expands to $3 \times 3$, triggering real-time softmax redistribution.
4. **Scene 04 — Separation**: Attention between Token A and Token B drops sharply (down to $0.198$) as attention shifts toward Token C ($0.401$). Kurva A and B separate vertically as their mutual bond weakens.
5. **Scene 05 — Aftermath**: All three curves reach steady state in their new contextual representations ($Output = \text{softmax}(QK^T / \sqrt{d_k}) V$).

---

## Getting Started

### Prerequisites
- Go 1.22 or higher
- Standard graphics driver (DirectX on Windows, OpenGL on Linux/macOS) — no CGO or external C/C++ compiler required.

### Running the Project
```bash
# 1. Download module dependencies
go mod tidy

# 2. Run numerical unit tests
go test -v .

# 3. Launch the application
go run .
```

To build a standalone executable:
```bash
go build -o attention_viz.exe .
.\attention_viz.exe
```

---

## Keyboard Controls & Interactions

You can interact with the visualization in real time using the keyboard or mouse:

| Key | Action |
| :--- | :--- |
| **`Space`** | Pause / resume animation |
| **`R`** | Restart from Scene 01 |
| **`1` – `5`** | Jump directly to Scene 1, 2, 3, 4, or 5 |
| **`←` / `→`** | Step to the previous / next scene |
| **`↑` / `↓`** or **`[` / `]`** | Adjust playback speed ($0.25\times$ to $3.0\times$) |
| **`M`** | Toggle the mathematical HUD / matrix inspector |
| **`C`** | Toggle causal masking ($M$ lower-triangular matrix) |
| **`H`** | Open the help overlay |
| **`Esc`** | Dismiss overlays |
| **Mouse Click** | Click any scene segment on the bottom timeline scrubber |

---

## Mathematical Foundation & Visual Mapping

The animation is driven directly by live matrix computations rather than arbitrary visual tweens:

- **Reference Matrices (from PDF EXP-019)**:
  $$Q = \begin{bmatrix} 1 & 0 \\ 0 & 1 \\ 1 & 1 \end{bmatrix}, \quad K = \begin{bmatrix} 1 & 0 \\ 0 & 1 \\ 1 & 1 \end{bmatrix}, \quad V = \begin{bmatrix} 1 & 0 \\ 0 & 2 \\ 2 & 1 \end{bmatrix}, \quad d_k = 2$$

- **Scores & Softmax**:
  Scores are calculated via $S = \frac{QK^T}{\sqrt{d_k}}$. Normalizing row-wise through Softmax yields weight matrix $\mathbf{A}$:
  $$\mathbf{A} \approx \begin{bmatrix} 0.4011 & 0.1978 & 0.4011 \\ 0.1978 & 0.4011 & 0.4011 \\ 0.2483 & 0.2483 & 0.5035 \end{bmatrix}$$

- **How Weights Control the Visuals**:
  1. **Waveform Coupling**: Token $i$'s curve inherits harmonic frequencies from Token $j$ weighted by $\alpha_{ij}$.
  2. **Trajectory Separation**: When Token C enters, $\alpha_{AB}$ drops from $0.3302$ to $0.1978$, reducing the mutual attraction between A and B and causing them to drift apart.
  3. **Filaments & Flux**: The beam width, brightness, and particle transfer rate along connecting filaments are directly proportional to $\alpha_{ij}$.
  4. **2D Subspace Plot**: The side HUD tracks original values $V$ shifting toward contextual outputs $O = \mathbf{A}V$.

---

## Project Structure

```text
├── main.go            # Window setup, game loop (60 FPS), and input dispatching
├── config.go          # Layout constants, color palette, scene metadata
├── attention.go       # Linear algebra engine (Q, K, V, Softmax, Causal Mask)
├── attention_test.go  # Unit tests verifying matrix math against PDF values
├── curves.go          # Harmonic wave generation and filament geometry
├── particles.go       # Particle physics for wave riders, attention packets, and ambient dust
├── simulation.go      # Bridges mathematical tensors with visual state over time
└── renderer.go        # Multi-pass vector drawing, glowing polylines, typography, HUD
```
