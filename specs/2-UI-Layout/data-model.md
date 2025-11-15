# Data Model: UI Layout and Layout Improvements

**Input**: Feature specification from `/specs/2-UI-Layout/layout_ui_improvements.md`

## Entities

### 1. Application State
**Purpose**: Main application model containing UI layout state
- **Fields**:
  - `ActiveTab` (TabType) - Current active tab (Timer, Tasks, Settings)
  - `WindowSize` (WindowSize) - Current terminal dimensions
  - `Styles` (Styles) - Lipgloss style definitions
  - `Tabs` ([]Tab) - List of available tabs
- **Relationships**: Contains references to all UI components
- **Validation**: ActiveTab must be one of available tabs
- **State Transitions**: ActiveTab changes when user selects different tab

### 2. WindowSize
**Purpose**: Represents terminal dimensions
- **Fields**:
  - `Width` (int) - Terminal width in characters
  - `Height` (int) - Terminal height in characters
- **Validation**: Width and Height must be positive
- **Relationships**: Referenced by layout components for responsive design

### 3. Tab
**Purpose**: Represents a single tab in the navigation system
- **Fields**:
  - `ID` (TabType) - Unique identifier (Timer, Tasks, Settings)
  - `Title` (string) - Display title for the tab
  - `Content` (string) - Rendered content for the tab
  - `IsActive` (bool) - Whether this tab is currently selected
- **Relationships**: Belongs to Application State
- **Validation**: ID must be unique among all tabs

### 4. Styles
**Purpose**: Container for all Lipgloss style definitions
- **Fields**:
  - `ActiveTabStyle` (lipgloss.Style) - Style for active tab
  - `InactiveTabStyle` (lipgloss.Style) - Style for inactive tabs
  - `HeaderStyle` (lipgloss.Style) - Style for header section
  - `ContentStyle` (lipgloss.Style) - Style for content area
  - `StatusBarStyle` (lipgloss.Style) - Style for status bar
  - `BorderStyle` (lipgloss.Style) - Style for borders
- **Relationships**: Used by all UI components for consistent styling
- **Validation**: All styles must be valid lipgloss.Style objects

### 5. LayoutConfiguration
**Purpose**: Defines responsive layout parameters
- **Fields**:
  - `HeaderHeight` (int) - Height of header section in characters
  - `ContentHeight` (int) - Height of content area in characters
  - `StatusBarHeight` (int) - Height of status bar in characters
  - `MinWidth` (int) - Minimum terminal width for proper display
  - `MinHeight` (int) - Minimum terminal height for proper display
- **Relationships**: Used by layout functions to calculate component sizes
- **Validation**: Heights must be positive, MinWidth/Height must be reasonable for UI

## State Transitions

### Tab Selection
- **From**: Any active tab
- **Trigger**: User selects different tab
- **To**: New active tab
- **Validation**: New tab must exist in available tabs list

### Window Resize
- **From**: Any layout state
- **Trigger**: Terminal window resized
- **To**: Updated layout based on new dimensions
- **Validation**: New dimensions must meet minimum requirements

### UI State Update
- **From**: Previous UI state
- **Trigger**: Application event (timer update, task change, etc.)
- **To**: Updated UI with new data
- **Validation**: Data integrity must be maintained

## Validation Rules

1. **Tab Consistency**: The active tab must always exist in the available tabs list
2. **Dimension Constraints**: UI components must adapt to terminal size constraints
3. **Style Consistency**: All UI elements must use defined styles for visual consistency
4. **Responsive Design**: Layout must remain functional across different terminal sizes
5. **Visual Hierarchy**: UI elements must follow the defined visual hierarchy for proper information display