# Quickstart Guide: UI Layout Improvements

**Input**: Feature specification from `/specs/2-UI-Layout/layout_ui_improvements.md`

## Overview

This guide will help you get started with the enhanced UI layout for the Pomodoro application. The new layout includes tab-based navigation, consistent styling using lipgloss, responsive design, and improved visual hierarchy.

## Prerequisites

- Go 1.21+ installed
- Basic understanding of Bubble Tea framework
- Understanding of lipgloss for styling and layout

## Setup

1. **Install Dependencies** (if not already installed)
   ```bash
   go mod tidy
   ```

2. **Update to the latest dependencies** (if needed)
   ```bash
   go get -u github.com/charmbracelet/bubbletea
   go get -u github.com/charmbracelet/lipgloss
   ```

## Key Components

### 1. Tab Navigation
- **Location**: `src/views/components/tab.go`
- **Purpose**: Allows switching between different views (Timer, Tasks, Settings)
- **Usage**: Use arrow keys or number keys to switch between tabs

### 2. Style Definitions
- **Location**: `src/views/styles.go`
- **Purpose**: Centralized lipgloss style definitions for consistent UI
- **Usage**: Import and apply styles to UI elements

### 3. Layout Utilities
- **Location**: `src/utils/layout.go`
- **Purpose**: Functions for responsive layout calculations
- **Usage**: Calculate component sizes based on terminal dimensions

### 4. Main View
- **Location**: `src/views/main_view.go`
- **Purpose**: Orchestrates the entire UI layout
- **Usage**: Renders header, content area, and status bar

## Running the Application

1. **Build the application**
   ```bash
   go build -o pomodoro cmd/main.go
   ```

2. **Run the application**
   ```bash
   ./pomodoro
   ```

3. **Using the enhanced UI**:
   - Press `Tab` or `Shift+Tab` to cycle through UI elements
   - Use arrow keys or number keys to switch between tabs (Timer, Tasks, Settings)
   - Press `q` or `Ctrl+C` to quit the application

## Making Changes

### Adding New Tabs
1. Define a new tab type in the `Application State` model
2. Create a view function to render the tab's content
3. Update the tab navigation component to include the new tab
4. Add the new tab to the tabs list in the application state

### Modifying Styles
1. Update style definitions in `src/views/styles.go`
2. Test the changes by running the application
3. Adjust values until desired appearance is achieved

### Responsive Design
1. Update layout calculations in `src/utils/layout.go`
2. Test in different terminal sizes
3. Adjust minimum dimensions if necessary

## Testing

1. **Unit Tests**: Run tests for individual components
   ```bash
   go test ./tests/unit/...
   ```

2. **Integration Tests**: Test UI component integration
   ```bash
   go test ./tests/integration/...
   ```

## Troubleshooting

### Layout Issues
- If layout appears broken, check terminal size requirements
- Ensure minimum width/height requirements are met
- Verify style definitions are properly applied

### Performance Issues
- If UI updates are slow, check for expensive rendering operations
- Optimize style calculations and caching where possible
- Verify efficient state updates

## Next Steps

1. Explore the `specs/2-UI-Layout/` directory for detailed specifications
2. Review the tasks in `specs/2-UI-Layout/tasks.md` for planned enhancements
3. Check the existing code in `src/views/` for implementation examples
4. Look at the component implementations in `src/components/` to understand how they integrate with the new layout