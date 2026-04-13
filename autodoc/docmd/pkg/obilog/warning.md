# `obilog` Package — Semantic Overview

The `obilog` package provides a lightweight, conditional logging interface for the OBItools4 ecosystem. It wraps `logrus`, a structured logger, to emit warnings only when explicitly allowed by application-wide settings.

## Core Functionality

- **`Warnf(format string, args ...interface{})`**  
  A convenience wrapper around `logrus.Warnf`, enabling formatted warning messages. It respects a global "silent warnings" toggle defined in `obidefault.SilentWarning()`.

## Design Intent

- **Conditional Warning Output**:  
  Warnings are suppressed when `obidefault.SilentWarning()` returns `true`, supporting quiet or batch execution modes (e.g., CI pipelines, automated runs).

- **Consistency & Integration**:  
  Leverages `obidefault` to enforce centralized control over verbosity, aligning logging behavior with higher-level application configuration.

- **Minimal Abstraction**:  
  Keeps the interface simple and idiomatic, avoiding over-engineering while preserving flexibility for future extensions (e.g., adding `Debugf`, `Infof` wrappers).

## Use Case

Ideal for non-fatal issues in command-line tools or libraries—where warnings should be visible by default but suppressible on demand, without altering core logic.

## Dependencies

- `logrus`: Structured logging backend  
- `obidefault`: Configuration layer for global behavior (e.g., silence mode)

> **Note**: This package is *not* a full logging subsystem—it’s a targeted, policy-aware warning emitter.
