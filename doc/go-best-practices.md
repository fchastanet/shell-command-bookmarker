# Go and TUI Best Practices

- [1. Overview](#1-overview)
- [2. Code easily testable](#2-code-easily-testable)
- [3. Function Design](#3-function-design)
- [4. Composition vs. Inheritance](#4-composition-vs-inheritance)
  - [4.1. Sample Composition Pattern](#41-sample-composition-pattern)

## 1. Overview

This document outlines best practices for Go applications. It covers Go-specific
design patterns.

## 2. Code easily testable

- Accept Interfaces, Return Structs
- Check `internal/services/ShellDetectionService.go` where every io functions
  are referenced as struct members, allowing for easy mocking in tests.

## 3. Function Design

- Constructor methods (`NewX`) should take structures as arguments rather than
  numerous parameters
- optional parameters should be passed using WithX functions
- Use interfaces to define component behaviors

## 4. Composition vs. Inheritance

Go does not support traditional inheritance or polymorphism. Instead:

- Use composition to reuse code
- Implement interfaces to define behavior
- Use the self-reference pattern for base models

### 4.1. Sample Composition Pattern

```go
type UIModel struct {
  Width  int
  Height int
  Self   any  // Store reference to actual implementation
}

type ChildModel struct {
  Base  *UIModel
  Field1 string
}

func NewChildModel() *ChildModel {
  childModel := &ChildModel{
    Base: &UIModel{
      Self: nil,  // Will be set later
    },
    Field1: "",
  }
  childModel.Base.Self = childModel  // Set self-reference
  return childModel
}
```
