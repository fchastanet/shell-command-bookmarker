# Go and TUI Best Practices

- [1. Overview](#1-overview)
- [2. Function Design](#2-function-design)
- [3. Composition vs. Inheritance](#3-composition-vs-inheritance)
  - [3.1. Sample Composition Pattern](#31-sample-composition-pattern)

## 1. Overview

This document outlines best practices for Go applications. It covers Go-specific
design patterns.

## 2. Function Design

- Constructor methods (`NewX`) should take structures as arguments rather than
  numerous parameters
- optional parameters should be passed using WithX functions
- Use interfaces to define component behaviors

## 3. Composition vs. Inheritance

Go does not support traditional inheritance or polymorphism. Instead:

- Use composition to reuse code
- Implement interfaces to define behavior
- Use the self-reference pattern for base models

### 3.1. Sample Composition Pattern

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
