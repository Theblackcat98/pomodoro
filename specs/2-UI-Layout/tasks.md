---

description: "Task list for UI Layout improvements feature"
---

# Tasks: UI Layout Improvements

**Input**: Design documents from `/specs/2-UI-Layout/`
**Prerequisites**: plan.md, research.md, layout_ui_improvements.md

**Tests**: Unit and integration tests will be implemented for each component to verify proper UI rendering and layout consistency.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `cmd/`, `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan in specs/2-UI-Layout/
- [ ] T002 [P] Update existing Go project with any necessary Lip Gloss dependencies in go.mod
- [ ] T003 [P] Set up testing framework for UI components in tests/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core UI infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Create style definitions in src/views/styles.go following lipgloss example patterns
- [ ] T005 [P] Create tab component using lipgloss styling in src/components/tab.go
- [ ] T006 [P] Create layout utility functions in src/utils/layout.go for responsive design
- [ ] T007 Update existing models to support new UI features in src/views/app.go
- [ ] T008 Define global style variables for consistent theming in src/views/styles.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enhanced Visual Structure (Priority: P1) 🎯 Layout MVP

**Goal**: Implement tab-based navigation system and organized layout structure with proper visual hierarchy

**Independent Test**: User can switch between Timer and Task views using tab navigation, with clear visual separation and organized layout

### Implementation for User Story 1

- [ ] T009 [P] [US1] Implement tab-based navigation in src/views/main_view.go
- [ ] T010 [P] [US1] Create header section with styled title in src/views/main_view.go
- [ ] T011 [US1] Implement main content area with organized layout for timer and tasks
- [ ] T012 [US1] Add status bar at bottom showing current status and controls
- [ ] T013 [US1] Implement proper visual hierarchy with borders, padding and margins
- [ ] T014 [US1] Write unit tests for new layout structure in tests/unit/layout_test.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Improved Styling & Themes (Priority: P2)

**Goal**: Implement consistent color palette, active/inactive states, and visual feedback

**Independent Test**: UI elements have proper styling with active/inactive states, consistent color palette, and visual separation

### Implementation for User Story 2

- [ ] T015 [P] [US2] Define consistent color palette in src/views/styles.go
- [ ] T016 [P] [US2] Implement active/inactive tab styles with visual feedback
- [ ] T017 [US2] Create consistent button and control styles
- [ ] T018 [US2] Add color definitions for different states (running, paused, etc.)
- [ ] T019 [US2] Implement visual elements like borders, padding, and margins for separation
- [ ] T020 [US2] Write unit tests for styling components in tests/unit/styling_test.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Responsive Layout & Adaptability (Priority: P3)

**Goal**: Ensure the UI adapts to different terminal sizes with proper alignment and sizing

**Independent Test**: UI properly adapts to different terminal dimensions while maintaining layout integrity

### Implementation for User Story 3

- [ ] T021 [P] [US3] Implement terminal size detection in src/views/app.go
- [ ] T022 [P] [US3] Create responsive layout calculations using lipgloss width detection
- [ ] T023 [US3] Implement proper alignment for different terminal sizes
- [ ] T024 [US3] Add sizing constraints that adapt to available space
- [ ] T025 [US3] Implement gap handling for different terminal widths
- [ ] T026 [US3] Write tests for responsive behavior in tests/unit/responsive_test.go

**Checkpoint**: At this point, User Stories 1, 2 AND 3 should all work together

---

## Phase 6: User Story 4 - Task View Enhancement (Priority: P4)

**Goal**: Enhance the task view with proper styling and layout similar to timer view

**Independent Test**: Task view has organized layout with styled components that match the new design

### Implementation for User Story 4

- [ ] T027 [P] [US4] Create styled task list component using lipgloss in src/components/task_list.go
- [ ] T028 [P] [US4] Implement task item styles with visual indicators
- [ ] T029 [US4] Add proper spacing and borders for task components
- [ ] T030 [US4] Integrate task list with tab navigation system
- [ ] T031 [US4] Write unit tests for enhanced task view in tests/unit/task_view_test.go

**Checkpoint**: At this point, all user stories should work together

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, polish, and bug fixes

- [ ] T032 [P] Review and refine all component styles for consistency
- [ ] T033 [P] Update documentation for new UI components
- [ ] T034 Test keyboard navigation throughout the new UI
- [ ] T035 Fix the issue where the timer doesn't start when pressing space
- [ ] T036 Fix the issue where pressing 'n' on task tab doesn't add a task
- [ ] T037 Performance testing of new UI components
- [ ] T038 Integration tests for complete UI flow

**Final Checkpoint**: Complete, tested, and polished UI layout implementation

---

## Dependencies

- **User Story 2** depends on foundational setup (Phase 2) being complete
- **User Story 3** depends on User Story 1 being complete
- **User Story 4** depends on User Story 1 being complete

## Parallel Execution Examples

- Tasks T005, T006 can run in parallel during Phase 2
- User stories 2, 3, and 4 can be developed in parallel after Phase 2 is complete
- All unit test tasks can be developed in parallel with their corresponding implementation tasks

## Implementation Strategy

1. **MVP Scope**: Complete User Story 1 (T009-T014) for basic tab-based navigation and layout
2. **Incremental Delivery**: Add styling (User Story 2), then responsiveness (User Story 3), then task view (User Story 4)
3. **Independent Testing**: Each user story can be tested independently for proper implementation
4. **Polish Phase**: Fix identified issues and optimize the complete UI