# `obilandmark` Package Overview

The `obilandmark` package provides semantic and programmatic access to landmark-related data, primarily for geospatial or augmented reality (AR) applications. It defines structured types and utilities to represent, query, and manage points of interest (POIs) with rich metadata.

## Core Functionalities

- **Landmark Representation**: Defines a `Landmark` struct with fields such as:
  - `ID`: Unique identifier (e.g., UUID or database key)
  - `Name`, `Description`, and optional categories/tags
  - Geocoordinates (`Latitude`, `Longitude`) with optional altitude & accuracy metadata

- **Metadata Enrichment**: Supports additional properties like:
  - Image URLs or embedded thumbnails
  - Opening hours, accessibility info (e.g., wheelchair-friendly)
  - Historical/cultural context or relevance flags

- **Geospatial Queries**: Offers functions to:
  - Filter landmarks within bounding boxes or radius-based regions
  - Sort by distance from a reference point (e.g., user location)
  - Handle coordinate transformations (WGS84, local projections)

- **Persistence & Sync**: Includes interfaces for:
  - Loading landmark datasets from JSON, GeoJSON, or SQLite
  - Incremental sync with remote APIs (e.g., OpenStreetMap extensions)

- **AR Integration Helpers**: Provides utilities for:
  - Calculating bearing/azimuth to a landmark relative to device orientation
  - Estimating visibility (e.g., line-of-sight, elevation masking)

- **Extensibility**: Designed for plugin-style extensions via interfaces (e.g., custom loaders, filters).

## Use Cases

- AR navigation apps
- Tourist guide systems  
- Smart city infrastructure overlays  
- Indoor/outdoor wayfinding

The package emphasizes semantic clarity, performance (via efficient indexing), and interoperability with standard geospatial formats.
