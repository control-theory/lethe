package graph

import (
	"math"
	"testing"
	"time"
)

func TestCurrentStrengthAt_NoDecay(t *testing.T) {
	now := time.Now().UnixMilli()
	// No time passed => no decay
	s := CurrentStrengthAt(1.0, now, 30, now)
	if math.Abs(s-1.0) > 0.001 {
		t.Errorf("strength = %v, want ~1.0", s)
	}
}

func TestCurrentStrengthAt_HalfLife(t *testing.T) {
	now := time.Now().UnixMilli()
	halfLifeDays := 30.0
	// Exactly 30 days ago => strength should be ~0.5
	thirtyDaysAgo := now - int64(30*24*time.Hour/time.Millisecond)
	s := CurrentStrengthAt(1.0, thirtyDaysAgo, halfLifeDays, now)
	if math.Abs(s-0.5) > 0.01 {
		t.Errorf("strength after 1 half-life = %v, want ~0.5", s)
	}
}

func TestCurrentStrengthAt_TwoHalfLives(t *testing.T) {
	now := time.Now().UnixMilli()
	sixtyDaysAgo := now - int64(60*24*time.Hour/time.Millisecond)
	s := CurrentStrengthAt(1.0, sixtyDaysAgo, 30, now)
	if math.Abs(s-0.25) > 0.01 {
		t.Errorf("strength after 2 half-lives = %v, want ~0.25", s)
	}
}

func TestCurrentStrengthAt_HigherInitial(t *testing.T) {
	now := time.Now().UnixMilli()
	thirtyDaysAgo := now - int64(30*24*time.Hour/time.Millisecond)
	s := CurrentStrengthAt(5.0, thirtyDaysAgo, 30, now)
	if math.Abs(s-2.5) > 0.05 {
		t.Errorf("strength 5.0 after 1 half-life = %v, want ~2.5", s)
	}
}

func TestReinforceBump(t *testing.T) {
	tests := []struct {
		current, bump, cap, want float64
	}{
		{1.0, 1.0, 10.0, 2.0},
		{9.5, 1.0, 10.0, 10.0}, // capped
		{0.0, 1.0, 10.0, 1.0},
		{10.0, 1.0, 10.0, 10.0}, // already at cap
	}
	for _, tt := range tests {
		got := ReinforceBump(tt.current, tt.bump, tt.cap)
		if got != tt.want {
			t.Errorf("ReinforceBump(%v, %v, %v) = %v, want %v", tt.current, tt.bump, tt.cap, got, tt.want)
		}
	}
}

func TestApplyDecay_Pinned(t *testing.T) {
	n := &Node{Strength: 1.0, Pinned: true, LastAccessedAt: 0}
	ApplyDecay(n, 30)
	if n.Strength != 1.0 {
		t.Errorf("pinned node decayed: strength = %v, want 1.0", n.Strength)
	}
}

func TestApplyDecay_Unpinned(t *testing.T) {
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour).UnixMilli()
	n := &Node{Strength: 1.0, Pinned: false, LastAccessedAt: thirtyDaysAgo}
	ApplyDecay(n, 30)
	if math.Abs(n.Strength-0.5) > 0.01 {
		t.Errorf("unpinned node after 30 days: strength = %v, want ~0.5", n.Strength)
	}
}
