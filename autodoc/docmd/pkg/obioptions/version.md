# OBIOptions Package – Semantic Description

The `obioptions` package provides a lightweight, version-aware utility for the OBITools suite. Its core functionality is centered around exposing runtime version information in a standardized and programmatic way.

## Key Features

- **Version Exposure**:  
  Exposes the current version of OBITools via a simple, read-only function `VersionString()`. This allows other modules or external tools to query the package version at runtime.

- **Automated Versioning**:  
  The `_Version` variable is automatically populated from an external `version.txt` file during the build process (via Makefile), ensuring consistency between source metadata and compiled artifacts.

- **Patch-Level Tracking**:  
  The version follows semantic conventions (`MAJOR.MINOR.PATCH`), with the patch number incremented automatically on each repository push—enabling precise tracking of development iterations.

- **No Side Effects**:  
  The `VersionString()` function is pure: it takes no parameters and performs only a string return, making it safe for use in logging, diagnostics, or compatibility checks.

- **Documentation Ready**:  
  Includes inline GoDoc comments for clarity and tooling support (e.g., `go doc`), improving maintainability.

## Use Cases

- Debugging and logging (e.g., including version in error reports).  
- Conditional logic based on OBITools compatibility.  
- CI/CD validation (e.g., verifying deployed version matches expectations).  

## Version Format

`"Release X.Y.Z"` (e.g., `"Release 4.4.29"`), where:
- `X` = Major release (breaking changes),
- `Y` = Minor release (new features, backward-compatible),
- `Z` = Patch level (incremented per push for hotfixes/bug fixes).

No external dependencies or configuration required.
