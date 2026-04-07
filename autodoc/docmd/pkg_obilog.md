# `obilog` Package — Semantic Overview

The `obilog` package provides a lightweight, conditional logging interface for the OBItools4 ecosystem. It wraps `logrus`, a structured logger, to emit warnings only when explicitly allowed by application-wide settings.

## Core Functionality

- **`Warnf(format string, args ...interface{})`**  
  Emits a formatted warning message using `logrus.Warnf`, subject to the global silence policy defined by `obidefault.SilentWarning()`. If warnings are silenced, this function becomes a no-op.

## Design Intent

- **Conditional Warning Output**:  
  Warnings are suppressed when `obidefault.SilentWarning()` returns `true`, supporting quiet or batch execution modes (e.g., CI pipelines, automated runs).

- **Consistency & Integration**:  
  Centralizes verbosity control via `obidefault`, ensuring logging behavior aligns with higher-level configuration without hardcoding logic.

- **Minimal Abstraction**:  
  Maintains a thin, idiomatic wrapper—avoiding over-engineering while preserving extensibility (e.g., future `Debugf`, `Infof` wrappers).

## Use Case

Designed for non-fatal issues in CLI tools or libraries—where warnings should be visible by default but suppressible on demand, *without* modifying core logic or sprinkling conditional checks throughout the codebase.

## Dependencies

- `logrus`: Structured logging backend (JSON/console formatting, hooks support)  
- `obidefault`: Configuration layer exposing global behavior flags (e.g., silence mode)

> **Note**: `obilog` is *not* a full logging subsystem—it’s a policy-aware warning emitter. It does **not** expose `Info`, `Debug`, or error-level logging; those should be handled directly via `logrus` where appropriate.
