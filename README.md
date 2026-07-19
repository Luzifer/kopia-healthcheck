# kopia-healthcheck

`kopia-healthcheck` monitors the maintenance state of a configured
[Kopia](https://kopia.io/) repository and reports the result to a
[Healthchecks.io](https://healthchecks.io/) ping URL.

On startup and at every configured interval, it runs:

```console
kopia maintenance info --json
```

It then checks that:

- the next quick-maintenance run is not overdue;
- the next full-maintenance run is not overdue;
- the latest run of every recorded maintenance task succeeded; and
- the configured maintenance owner matches an expected owner, when enforced.

A successful result is posted to the configured URL. If any check fails, the
same report is posted to the URL with `/fail` appended. Each request body
contains a human-readable summary of the individual checks.

## Requirements

- Access to a `kopia` executable
- Access to an already configured and connected Kopia repository
- A Healthchecks.io ping URL

The healthcheck process must run with the same Kopia configuration and
credentials that are used to access the repository. Do not pass repository
passwords or other secrets as command-line arguments where they may be exposed
through the process list.

## Usage

Build and run the binary:

```console
go build -o kopia-healthcheck .
./kopia-healthcheck --healthcheck-url=https://hc-ping.com/your-check-uuid
```

Configuration flags can also be supplied through uppercase environment
variables with dashes replaced by underscores, for example:

```console
HEALTHCHECK_URL=https://hc-ping.com/your-check-uuid \
CHECK_INTERVAL=30m \
RUN_GRACE=10m \
./kopia-healthcheck
```

Available options:

| Flag | Default | Description |
| --- | --- | --- |
| `--allow-disabled-runs` | `false` | Treat disabled quick or full maintenance as healthy |
| `--check-interval` | `30m` | Time between checks |
| `--expected-owner` | empty | Require the configured Kopia maintenance owner to match this value |
| `--healthcheck-url`, `-u` | empty | Healthchecks.io ping URL |
| `--log-level` | `info` | Log level: `debug`, `info`, `warn`, `error`, or `fatal` |
| `--run-grace` | `10m` | Time an overdue next-run timestamp is tolerated |
| `--version` | `false` | Print the version and exit |

Durations use Go duration syntax, such as `30s`, `10m`, or `24h`.

## Container

The included multi-stage `Dockerfile` builds the application, copies it into
the upstream Kopia image, and configures it as the container entrypoint:

```console
docker build -t kopia-healthcheck .
docker run --rm \
  -e HEALTHCHECK_URL=https://hc-ping.com/your-check-uuid \
  kopia-healthcheck
```

Mount the Kopia configuration and provide its repository credentials using the
same mechanism as your existing Kopia container. The exact mounts and variables
depend on how that repository was configured.

## Failure behavior

- A failed maintenance-state check sends a Healthchecks.io failure ping.
- Failure to execute Kopia, decode its output, or contact Healthchecks.io is
  logged and retried at the next interval.
- Each complete check has a one-minute timeout shared by the Kopia command and
  the HTTP request.
- Disabled quick or full maintenance is considered a failure unless
  `--allow-disabled-runs` is enabled.
