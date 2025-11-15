# Quickstart Guide: Pomodoro TUI App

## Getting Started

This guide will help you get the Pomodoro TUI application up and running quickly.

### Prerequisites

- Go 1.21 or higher
- Git
- A terminal that supports UTF-8 characters and ANSI colors

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd pomodoro-app
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o pomodoro cmd/main.go
   ```

4. Run the application:
   ```bash
   ./pomodoro
   ```

### Running in Development Mode

For development, you can run the application directly:

```bash
go run cmd/main.go
```

## Using the Application

### Main Screen

When you start the application, you'll see:
- A large circular timer showing remaining time
- Current mode indicator ("Focus" or "Break")
- A central START/PAUSE button
- Duration selection capsules (15, 25, 35 min)

### Starting a Session

1. Adjust the duration if needed by selecting a capsule
2. Press the START button in the center
3. The timer will begin counting down
4. The button will change to PAUSE

### Pausing/Resuming

- Press PAUSE to stop the timer
- Press START again to resume from where you left off

### Navigation

The application uses keyboard navigation:

| Key | Action |
|-----|--------|
| Space | Start/Pause timer |
| Enter | Confirm actions |
| ←/→ | Navigate selection |
| ↑/↓ | Move between components |
| Tab | Cycle between UI components |
| Esc | Return to main screen |
| h | Show help |
| q | Quit application |
| s | Switch to stopwatch mode |
| t | Open tasks view |
| d | Open settings |

### Available Components

This application showcases all Bubbles components integrated meaningfully:

1. **Timer Component**: Main Pomodoro timer functionality
2. **Stopwatch Component**: Alternative timing mode
3. **Progress Component**: Visual time representation
4. **List Component**: Task management and history
5. **Text Input Component**: Settings and task entry
6. **Text Area Component**: Task descriptions
7. **Spinner Component**: Loading indicators
8. **Help Component**: Contextual help
9. **Key Component**: Customizable key bindings
10. **Table Component**: Session statistics
11. **File Picker Component**: Import/export functionality
12. **Paginator Component**: Paginated history
13. **Viewport Component**: Scrollable content areas
14. **Glow Component**: Markdown help content

### Settings

Access settings using the 'd' key or settings button:

- Customize focus/break durations
- Set notification preferences (sound, vibration, silent)
- Select color themes
- Configure auto-start options

### Tasks

Track your work with the task system:

1. Press 't' to access task view
2. Add new tasks with titles and descriptions
3. Track how many Pomodoro sessions each task takes
4. Mark tasks as complete when finished

## Development

### Project Structure

```
pomodoro/
├── cmd/
│   └── main.go           # Application entry point
├── src/
│   ├── models/           # Data structures
│   ├── services/         # Business logic
│   ├── components/       # Bubbles UI components
│   └── views/            # Combined UI views
├── specs/                # Specification documents
│   └── 1-main-timer/     # Current feature specs
└── docs/                 # Documentation
```

### Adding New Components

When adding new functionality:

1. Create a new component in `src/components/`
2. Follow the Bubble Tea pattern with Model, Init, Update, and View functions
3. Style with Lipgloss for consistency
4. Add to the main View composition
5. Ensure it integrates with existing state management

### Testing

Run the test suite:

```bash
go test ./...
```

Add tests for new functionality in the respective component directories.

## Troubleshooting

### Terminal Display Issues

If the UI appears misaligned:

1. Ensure your terminal supports UTF-8
2. Try resizing your terminal window
3. Check that your terminal supports ANSI colors

### Performance Issues

If the application is slow:

1. Check your terminal's performance (some terminals are slower than others)
2. Ensure you have the latest Go version installed
3. Close other applications to free up system resources

## Next Steps

1. Explore the task management features
2. Customize your Pomodoro settings
3. Try the statistics view to track your productivity
4. Experiment with different themes