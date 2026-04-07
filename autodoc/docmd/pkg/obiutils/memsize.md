# `obiutils` Package: Memory Size Parsing and Formatting

This Go package provides two complementary utility functions for handling human-readable memory sizes:

- **`ParseMemSize(s string) (int, error)`**  
  Parses a memory size string into an integer number of bytes. Supports case-insensitive units: `B`, `K`/`KB`, `M`/`MB`, `G`/`GB`, and `T`/`TB`.  
  Examples: `"128K"` → `131072`, `"512MB"` → `536870912`.  
  Returns an error for invalid input (e.g., empty string, non-numeric prefix, or unknown unit).

- **`FormatMemSize(n int) string`**  
  Converts a byte count into the most appropriate human-readable format using powers of 1024.  
  Uses suffixes `T`, `G`, `M`, or `K`; falls back to bytes (`B`) if < 1 KiB.  
  Integers are displayed without decimals (e.g., `2048` → `"2K"`), while fractional values use one decimal (e.g., `1536` → `"1.5K"`).

Both functions ensure semantic clarity and consistency for memory-related I/O, logging, or configuration parsing.
