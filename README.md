# 🍅 Pomodoro TUI

A terminal-based Pomodoro timer built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

Focus and break sessions follow the classic rhythm — focus, short break, and a
long break every few rounds — with a progress ring, task tracking, session
history, and settings that survive restarts.

## Screenshots

### Timer

![Timer — ready](docs/screenshots/timer.png)

![Timer — running](docs/screenshots/timer-running.png)

The progress ring sweeps clockwise in the session's color — tomato for focus,
green for short breaks, blue for long breaks — around large digits, with the
current task underneath.

### Tasks

![Tasks](docs/screenshots/tasks.png)

Tasks carry 🍅 estimates vs. completed pomodoros; focusing a task links your
sessions to it. `a` opens a slim add-bar with title, description, and estimate.

### Stats

![Stats](docs/screenshots/stats.png)

Today's pomodoros and focus time, the last 7 days, and recent sessions with
task attribution.

### Settings

![Settings](docs/screenshots/settings.png)

Durations, long-break interval, auto-start behavior, notifications, and an
opt-in mouse mode (wheel scrolling + clickable tabs) — every change applies and
persists immediately.

### Help

![Help](docs/screenshots/help.png)

A scrollable cheat sheet, available anywhere via `?`.

Quitting while a session runs asks for confirmation first:

![Quit confirmation](docs/screenshots/quit-confirm.png)

## Getting started

```bash
go build -o pomodoro ./cmd
./pomodoro
```

Or run it directly:

```bash
go run ./cmd
```

State (tasks, settings, session history) is stored as JSON in
`~/.config/pomodoro/state.json`.

## Keyboard shortcuts

| Key              | Action                                    |
| ---------------- | ----------------------------------------- |
| `space` / `enter`| Start, pause, resume the session          |
| `r`              | Reset to ready (abandons the session)     |
| `s`              | Skip to the next session                  |
| `1` / `2` / `3`  | Quick-start a 15 / 25 / 35 minute focus   |
| `tab`            | Next view (Timer → Tasks → Stats → …)     |
| `shift+tab`      | Previous view                             |
| `a`              | Add a task (Tasks view)                   |
| `x`              | Toggle a task done                        |
| `d`              | Delete a task (`y` confirms)              |
| `enter`          | Focus the selected task                   |
| `j` / `k`        | Move through lists and panels             |
| `+` / `-`        | Adjust the selected setting (Settings)    |
| `?`              | Help                                      |
| `q`              | Quit (asks for confirmation while running)|
| `esc`            | Cancel / go back                          |

## Repository layout

- `cmd/` — program entry point
- `src/views/` — views, styles, and the root model
- `src/components/` — reusable Bubble Tea components (list, inputs, ring, …)
- `src/models/` — data models and messages
- `src/services/` — timer engine and persistence
- `specs/` — feature specifications and plans
- `docs/` — research notes and the design vision (`docs/layout.md`)
