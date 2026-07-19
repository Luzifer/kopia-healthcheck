package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"git.luzifer.io/luzifer/kopia-healthcheck/pkg/kopia"
)

type (
	checkFn func(kopia.MaintenanceInfo) (title string, success bool, err error)
)

func checkNextFullInFuture(m kopia.MaintenanceInfo) (title string, success bool, err error) {
	if !m.Full.Enabled {
		return "Full-Runs are Disabled", cfg.AllowDisabledRuns, nil
	}

	title = "Next Full-Run in Future"
	success = m.Schedule.NextFullMaintenance.Add(cfg.RunGrace).After(time.Now())
	return title, success, nil
}

func checkLastRunSuccess(m kopia.MaintenanceInfo) (title string, success bool, err error) {
	var keysErrored []string
	for key, runs := range m.Schedule.Runs {
		if len(runs) < 1 {
			continue
		}

		sort.Slice(runs, func(i, j int) bool {
			return runs[i].Start.After(runs[j].Start)
		})

		if !runs[0].Success {
			keysErrored = append(keysErrored, key)
		}
	}

	if len(keysErrored) == 0 {
		return "All Latest Task-Runs Successful", true, nil
	}

	sort.Strings(keysErrored)

	return fmt.Sprintf("Tasks errored: %s", strings.Join(keysErrored, ", ")), false, nil
}

func checkNextQuickInFuture(m kopia.MaintenanceInfo) (title string, success bool, err error) {
	if !m.Quick.Enabled {
		return "Quick-Runs are Disabled", cfg.AllowDisabledRuns, nil
	}

	title = "Next Quick-Run in Future"
	success = m.Schedule.NextQuickMaintenance.Add(cfg.RunGrace).After(time.Now())
	return title, success, nil
}

func checkOwner(m kopia.MaintenanceInfo) (title string, success bool, err error) {
	if cfg.ExpectedOwner == "" {
		return "Maintenance Owner is not enforced", true, nil
	}

	if cfg.ExpectedOwner == m.Owner {
		return "Maintenance Owner Matches", true, nil
	}

	return fmt.Sprintf("Maintenance Owner is set to %q", m.Owner), false, nil
}
