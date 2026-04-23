package config

import (
	"context"
	"database/sql"
	"strconv"
)

// DecayConfig holds the tunable decay parameters.
type DecayConfig struct {
	HalfLifeDays      float64
	ReinforcementBump float64
	StrengthCap       float64
}

// DefaultDecayConfig returns the default decay parameters.
func DefaultDecayConfig() DecayConfig {
	return DecayConfig{
		HalfLifeDays:      30,
		ReinforcementBump: 1.0,
		StrengthCap:       10.0,
	}
}

// LoadDecayConfig reads decay parameters from the config table.
func LoadDecayConfig(ctx context.Context, db *sql.DB) (DecayConfig, error) {
	cfg := DefaultDecayConfig()

	rows, err := db.QueryContext(ctx, `SELECT key, value FROM config WHERE key IN ('half_life_days', 'reinforcement_bump', 'strength_cap')`)
	if err != nil {
		return cfg, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return cfg, err
		}
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		switch key {
		case "half_life_days":
			cfg.HalfLifeDays = f
		case "reinforcement_bump":
			cfg.ReinforcementBump = f
		case "strength_cap":
			cfg.StrengthCap = f
		}
	}
	return cfg, rows.Err()
}

// SaveDecayConfig writes decay parameters to the config table.
func SaveDecayConfig(ctx context.Context, db *sql.DB, cfg DecayConfig) error {
	for _, kv := range []struct {
		key string
		val float64
	}{
		{"half_life_days", cfg.HalfLifeDays},
		{"reinforcement_bump", cfg.ReinforcementBump},
		{"strength_cap", cfg.StrengthCap},
	} {
		_, err := db.ExecContext(ctx,
			`INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?`,
			kv.key, strconv.FormatFloat(kv.val, 'f', -1, 64), strconv.FormatFloat(kv.val, 'f', -1, 64))
		if err != nil {
			return err
		}
	}
	return nil
}
