package graph

import (
	"math"
	"time"
)

const millisPerDay = float64(24 * time.Hour / time.Millisecond)

// CurrentStrength computes the decay-adjusted strength of a node.
// S(t) = S₀ × e^(-λ × Δt)
// where λ = ln(2) / halfLifeDays and Δt is in days.
func CurrentStrength(storedStrength float64, lastAccessedAtMs int64, halfLifeDays float64) float64 {
	if halfLifeDays <= 0 {
		return storedStrength
	}
	lambda := math.Log(2) / halfLifeDays
	deltaDays := float64(time.Now().UnixMilli()-lastAccessedAtMs) / millisPerDay
	if deltaDays < 0 {
		deltaDays = 0
	}
	return storedStrength * math.Exp(-lambda*deltaDays)
}

// CurrentStrengthAt computes decay-adjusted strength at a specific time.
// Used for testing with deterministic timestamps.
func CurrentStrengthAt(storedStrength float64, lastAccessedAtMs int64, halfLifeDays float64, nowMs int64) float64 {
	if halfLifeDays <= 0 {
		return storedStrength
	}
	lambda := math.Log(2) / halfLifeDays
	deltaDays := float64(nowMs-lastAccessedAtMs) / millisPerDay
	if deltaDays < 0 {
		deltaDays = 0
	}
	return storedStrength * math.Exp(-lambda*deltaDays)
}

// ReinforceBump computes the new strength after a reinforcement event.
// Capped at cap to prevent runaway.
func ReinforceBump(currentStrength, bump, cap float64) float64 {
	s := currentStrength + bump
	if s > cap {
		return cap
	}
	return s
}

// ApplyDecay adjusts a node's strength in place based on decay.
// Pinned nodes are not decayed.
func ApplyDecay(n *Node, halfLifeDays float64) {
	if n.Pinned {
		return
	}
	n.Strength = CurrentStrength(n.Strength, n.LastAccessedAt, halfLifeDays)
}
