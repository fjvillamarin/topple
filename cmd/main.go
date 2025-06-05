// main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/alecthomas/kong"
)

var Version = "dev" // This will be set by the build system
type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type Globals struct {
	Debug     bool        `help:"Enable debug logging" short:"d"`
	Version   VersionFlag `name:"version" help:"Print version information and quit"`
	Recursive bool        `help:"Process directories recursively" short:"r"`
	TSLib     string      `help:"Path to the Tree-sitter library binary" short:"t" default:"./tree-sitter-sylfie/sylfie.dylib"`
}

// CLI holds the root command structure including global flags
type CLI struct {
	Globals

	// Commands
	Compile CompileCmd `cmd:"" help:"Compile PSX files to Python"`
	Watch   WatchCmd   `cmd:"" help:"Watch for changes and compile on the fly"`
	Scan    ScanCmd    `cmd:"" help:"Run the scanner and show/output tokens"`
	Parse   ParseCmd   `cmd:"" help:"Parse source files and show/output AST"`
}

func main() {
	// -------------------------------------------------------------------------
	// Parse CLI arguments and options
	cli := CLI{}

	// If no arguments are provided, show help
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}

	// Parse the command line arguments
	kCtx := kong.Parse(&cli,
		kong.Name("sylfie"),
		kong.Description("Sylfie Compiler CLI - compile PSX views into Python code"),
		kong.UsageOnError(),
		kong.Vars{
			"version": "v0.1.0",
		},
	)

	// -------------------------------------------------------------------------
	// Logger
	level := slog.LevelInfo

	if cli.Globals.Debug {
		level = slog.LevelDebug
	}

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}),
	)

	// -------------------------------------------------------------------------
	// Context

	ctx := context.Background()

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.DebugContext(ctx, "startup", slog.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)))

	// -------------------------------------------------------------------------
	// Run

	if err := kCtx.Run(&cli.Globals, &ctx, log); err != nil {
		kCtx.FatalIfErrorf(err)
	}
}
