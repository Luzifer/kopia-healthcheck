// Monitor kopia maintenance runs
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Luzifer/rconfig/v2"
	"github.com/sirupsen/logrus"

	"git.luzifer.io/luzifer/kopia-healthcheck/pkg/kopia"
)

var (
	cfg = struct {
		AllowDisabledRuns bool          `flag:"allow-disabled-runs" default:"false" description:"Do not report disabled quick/full-runs as issue"`
		CheckInterval     time.Duration `flag:"check-interval" default:"30m" description:"How often to check the maintenance status"`
		ExpectedOwner     string        `flag:"expected-owner" default:"" description:"When set checks whether the given owner is set for maintenance"`
		HealthcheckURL    string        `flag:"healthcheck-url,u" default:"" description:"Healthchecks.io URL to ping with status"`
		LogLevel          string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		RunGrace          time.Duration `flag:"run-grace" default:"10m" description:"How long to accept next-run to be in the past"`
		VersionAndExit    bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func initApp() error {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		return fmt.Errorf("parsing cli options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	if cfg.CheckInterval <= 0 {
		return fmt.Errorf("check-interval must be a positive duration")
	}

	if cfg.RunGrace < 0 {
		return fmt.Errorf("run-grace must be a duration greater or equal zero")
	}

	if _, err = url.Parse(cfg.HealthcheckURL); err != nil {
		return fmt.Errorf("parsing healthcheck-url: %w", err)
	}

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		logrus.WithField("version", version).Info("kopia-healthcheck")
		os.Exit(0)
	}

	logrus.WithField("version", version).Info("kopia-healthcheck started")

	if err = checkMaintenance(); err != nil {
		logrus.WithError(err).Error("checking maintenance (initial)")
	}

	checkTimer := time.NewTicker(cfg.CheckInterval)

	for range checkTimer.C {
		if err = checkMaintenance(); err != nil {
			logrus.WithError(err).Error("checking maintenance")
		}
	}
}

func checkMaintenance() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	logrus.Debug("starting check...")

	info, err := kopia.GetMaintenanceInfo(ctx)
	if err != nil {
		return fmt.Errorf("getting maintenance info: %w", err)
	}

	var (
		results []string
		success = true
	)
	for _, fn := range []checkFn{
		checkNextQuickInFuture,
		checkNextFullInFuture,
		checkLastRunSuccess,
		checkOwner,
	} {
		title, checkSuccess, err := fn(info)
		if err != nil {
			return fmt.Errorf("checking %q: %w", title, err)
		}

		logrus.WithFields(logrus.Fields{
			"success": checkSuccess,
			"title":   title,
		}).Debug("sub-check completed")

		if checkSuccess {
			results = append(results, fmt.Sprintf("✅ %s", title))
		} else {
			results = append(results, fmt.Sprintf("⛔ %s", title))
			success = false
		}
	}

	logrus.WithField("success", success).Info("check completed")

	checkURL := cfg.HealthcheckURL
	if !success {
		checkURL = strings.Join([]string{strings.TrimRight(checkURL, "/"), "fail"}, "/")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkURL, strings.NewReader(strings.Join(results, "\n")))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("closing response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
