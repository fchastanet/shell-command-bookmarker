# TUI Best Practices

- [1. Overview](#1-overview)
- [2. Project Architecture](#2-project-architecture)
- [3. TUI Best Practices](#3-tui-best-practices)
  - [3.1. Component Hierarchy](#31-component-hierarchy)
  - [3.2. UI Consistency](#32-ui-consistency)
  - [3.3. Hierarchical Message Flow](#33-hierarchical-message-flow)
  - [3.4. Event Propagation Patterns](#34-event-propagation-patterns)
  - [3.5. Command Management](#35-command-management)
- [4. Base Model implementation suggestion](#4-base-model-implementation-suggestion)
  - [4.1. UIModel Pattern](#41-uimodel-pattern)
  - [4.2. Event Handling Examples](#42-event-handling-examples)
    - [4.2.1. Example 1: Window Size Propagation](#421-example-1-window-size-propagation)
  - [4.3. Example 2: Key Event Handling](#43-example-2-key-event-handling)
- [5. Resources](#5-resources)

## 1. Overview

This document outlines best practices for developing Terminal User Interface
(TUI) applications in Go, with a focus on the BubbleTea framework. It covers
component architecture, event handling and message flow.

## 2. Project Architecture

- `internal/models`: Contains all TUI models
  - `top`: Main TUI model that instantiates all child models using maker
    functions
    - Instantiated models are stored in cache
    - `top` inherits from `internal/models/pane_manager.go`
  - Children models:
    - `explorer`: Navigation through folders
    - `command`: Command listing and editing
  - Shared packages:
    - `styles`: Styling information shared across models
    - `keys`: Keybinding information shared across models

## 3. TUI Best Practices

### 3.1. Component Hierarchy

- Use sub-models instead of modes for different UI states
- Each component should have a single responsibility

### 3.2. UI Consistency

- Global keymap with component-specific defaults
  - Binding names should be based on features (not keys)
  - Global keybinding help system
- Global style system with component-specific defaults

### 3.3. Hierarchical Message Flow

- Messages flow in logical priority order
- Special message types handled first
- Generic handling as fallback

### 3.4. Event Propagation Patterns

- **Descendant Flow** (parent → child):
  - Window size events
  - Focus/blur events
- **Ascendant Flow** (child → parent):
  - Key events (focused elements only)
- **Mixed Flow**:
  - Global keys (quit, help) at top level
  - Component-specific keys (navigation) at component level

### 3.5. Command Management

- Background event relaying
  ([example](https://github.dev/leg100/pug/internal/tui/top/start.go))
- The handler of a `tea.Cmd` cannot return a `tea.Cmd`, but must return
  `tea.Msg` instead

## 4. Base Model implementation suggestion

### 4.1. UIModel Pattern

Every TUI component could:

1. Use `UIModel` as a base model
2. Implement `Update` method that calls the base model's `Update`
3. Handle specific events through dedicated methods

```go
// Base model handles common events through interfaces
func (a *UIModel) Update(msg tea.Msg) tea.Cmd {
  switch msg := msg.(type) {
    case tea.WindowSizeMsg:
      if handler, ok := a.Self.(WindowSizeInterface); ok {
        return handler.HandleWindowResize(msg)
      }
    case FocusMsg:
      if handler, ok := a.Self.(FocusableInterface); ok {
        return handler.Focus()
      }
    // Other common event types...
  }
  return nil
}
```

### 4.2. Event Handling Examples

#### 4.2.1. Example 1: Window Size Propagation

```text
topPaneModel
  -> Update (WindowSizeMsg)
  -> topPaneModel.HandleWindowResize (WindowSizeMsg)
      -> childModel.Update (WindowSizeMsg)
      -> childModel.HandleWindowResize (WindowSizeMsg)
        -> ...
```

### 4.3. Example 2: Key Event Handling

```text
topPaneModel
  -> Update (KeyMsg)
  -> topPaneModel.HandleKeyMsg (KeyMsg)
    -> switch with keys that need to be handled at ascendant level (Quit, ...)
    -> if key not handled or propagation not stopped
      -> childModel.Update (KeyMsg)
      -> childModel.HandleKeyMsg (KeyMsg)
    -> if cmd with propagation not stopped
      -> switch with keys that need to be handled at descendant level
```

## 5. Resources

- [Building BubbleTea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/)
