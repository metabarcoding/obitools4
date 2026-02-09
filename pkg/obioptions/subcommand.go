package obioptions

import (
	"github.com/DavidGamba/go-getoptions"
)

// GenerateSubcommandParser creates an option parser that supports subcommands
// via go-getoptions' NewCommand/SetCommandFn/Dispatch API.
//
// The setup function receives the root *GetOpt and should register subcommands
// using opt.NewCommand(). Global options (--debug, --max-cpu, etc.) are
// registered before setup is called and are inherited by all subcommands.
//
// Returns the root *GetOpt (needed for Dispatch) and an ArgumentParser
// that handles parsing and post-parse processing.
func GenerateSubcommandParser(
	program string,
	documentation string,
	setup func(opt *getoptions.GetOpt),
) (*getoptions.GetOpt, ArgumentParser) {

	options := getoptions.New()
	options.Self(program, documentation)
	options.SetMode(getoptions.Bundling)
	options.SetUnknownMode(getoptions.Fail)

	// Register global options (inherited by all subcommands)
	RegisterGlobalOptions(options)

	// Let the caller register subcommands
	setup(options)

	// Add automatic help subcommand (must be after all commands)
	options.HelpCommand("help", options.Description("Show help for a command"))

	parser := func(args []string) (*getoptions.GetOpt, []string) {
		remaining, err := options.Parse(args[1:])
		ProcessParsedOptions(options, err)
		return options, remaining
	}

	return options, parser
}
