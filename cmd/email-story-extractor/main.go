package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, writer io.Writer) error {
	flags := flag.NewFlagSet("email-story-extractor", flag.ContinueOnError)
	flags.SetOutput(writer)

	maildir := flags.String("maildir", "", "path to the Maildir")
	config := flags.String("config", "", "path to the TOML configuration file")
	verbose := flags.Bool("verbose", false, "enable DEBUG-level logs")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *maildir == "" || *config == "" {
		flags.Usage()
		return fmt.Errorf("maildir and config flags are required")
	}

	output := zerolog.ConsoleWriter{
		Out:        writer,
		NoColor:    true,
		TimeFormat: time.RFC3339,
	}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("level=%s", i)
	}
	logger := zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger.Info().Msg("starting email story extraction")
	logger.Debug().Msg("this is a debug message")

	return nil
}
