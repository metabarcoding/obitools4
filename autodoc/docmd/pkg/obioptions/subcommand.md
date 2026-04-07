# Semantic Description of `GenerateSubcommandParser`

The function `GenerateSubcommandParser` constructs a command-line argument parser with support for **subcommands**, leveraging the `go-getoptions` library.

- It accepts:
  - `program`: The program name (used for help/version).
  - `documentation`: A top-level description of the tool.
  - `setup`: A callback to register subcommands and their options.

- Internally:
  - Initializes a `GetOpt` instance with bundling mode (`-abc`) and strict unknown-option handling.
  - Registers **global options** (e.g., `--debug`, `--verbose`) that are inherited by all subcommands.
  - Invokes the user-provided `setup` function to define **subcommand-specific options and commands**.
  - Automatically adds a built-in `help` subcommand for command-level documentation.

- Returns:
  - The root `*GetOpt`, required to invoke `.Dispatch()`.
  - An `ArgumentParser` function (signature: `func([]string) (*GetOpt, []string)`), which:
    - Parses command-line arguments (skipping `args[0]`, typically the binary name),
    - Handles errors via `ProcessParsedOptions`,
    - Returns parsed state and remaining positional arguments.

This design enables a clean, hierarchical CLI structure: global flags → subcommands → per-command options/positional args.
