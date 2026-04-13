# `obiutils` Package — Semantic Feature Summary

This Go package provides a set of utility functions for **type conversion**, **casting validation**, and **map/slice transformation** in a flexible, error-tolerant manner.

- `InterfaceToString(i interface{})`: Converts any value to a string, preferring the `Stringer` interface if implemented.
- `CastableToInt(i interface{})`: Checks whether a value is *numerically castable* to an `int` (supports all numeric types).
- `InterfaceToBool(i interface{})`: Safely converts various input types (`bool`, numeric, string like `"true"`, `"1"`, etc.) to `bool`; returns a custom error for unsupported types.
- `InterfaceToInt(i interface{})`: Converts numeric or string representations to an `int`, with precise error handling.
- `InterfaceToFloat64(i interface{})`: Converts numeric or string types to `float64`, using standard parsing.
- `MapToMapInterface(m interface{})`: Converts specialized map types (e.g., read-only or concurrency-safe maps) to `map[string]interface{}` via reflection.
- `InterfaceToIntMap(i interface{})`: Converts compatible map types (`map[string]int`, `hasMap` interfaces, or generic maps) to a concrete `map[string]int`.
- `InterfaceToStringMap(i interface{})`: Converts map values to strings, yielding a clean `map[string]string`.
- `InterfaceToStringSlice(i interface{})`: Converts slices of interfaces or strings into a pure `[]string`.

All functions include **explicit error handling** via custom types (e.g., `NotAnInteger`, `NotAMapInt`) and use logging via Logrus for debugging. The package prioritizes **type safety**, **robustness**, and interoperability across Go types.
