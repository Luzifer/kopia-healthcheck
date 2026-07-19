// Package kopia provides access to maintenance information reported by Kopia.
package kopia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type (
	// MaintenanceInfo contains the configured maintenance plans and their schedule.
	MaintenanceInfo struct {
		Owner string          `json:"owner"`
		Quick MaintenancePlan `json:"quick"`
		Full  MaintenancePlan `json:"full"`

		Schedule struct {
			NextFullMaintenance  time.Time                   `json:"nextFullMaintenance"`
			NextQuickMaintenance time.Time                   `json:"nextQuickMaintenance"`
			Runs                 map[string][]MaintenanceRun `json:"runs"`
		} `json:"schedule"`
	}

	// MaintenancePlan describes the configuration of a maintenance cycle.
	MaintenancePlan struct {
		Enabled  bool          `json:"enabled"`
		Interval time.Duration `json:"interval"`
	}

	// MaintenanceRun describes the outcome of a maintenance task execution.
	MaintenanceRun struct {
		Start   time.Time `json:"start"`
		End     time.Time `json:"end"`
		Success bool      `json:"success"`
		Error   string    `json:"error,omitempty"`
	}
)

// GetMaintenanceInfo executes Kopia and returns its current maintenance information.
func GetMaintenanceInfo(ctx context.Context) (m MaintenanceInfo, err error) {
	buf := new(bytes.Buffer)

	cmd := exec.CommandContext(ctx, "kopia", "maintenance", "info", "--json")

	cmd.Env = os.Environ()
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return m, fmt.Errorf("running kopia command: %w", err)
	}

	if err = json.NewDecoder(buf).Decode(&m); err != nil {
		return m, fmt.Errorf("decoding output: %w", err)
	}

	return m, nil
}
