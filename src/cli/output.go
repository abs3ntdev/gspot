package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Format represents an output format for CLI responses.
type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
	FormatSilent Format = "silent"
)

// getFormat reads the --format/--json flags from the command.
func getFormat(cmd *cli.Command) Format {
	if cmd.Bool("json") {
		return FormatJSON
	}
	f := cmd.String("format")
	switch Format(f) {
	case FormatJSON, FormatSilent, FormatPretty:
		return Format(f)
	default:
		return FormatPretty
	}
}

// getWriter returns the output writer based on --output flag.
func getWriter(cmd *cli.Command) (io.WriteCloser, error) {
	path := cmd.String("output")
	if path == "" || path == "-" {
		return os.Stdout, nil
	}
	return os.Create(path)
}

// Output renders a proto message according to the given format.
// The pretty function is the type-specific formatter for pretty mode.
// JSON and silent work generically on any proto.Message.
func Output[T proto.Message](format Format, w io.Writer, msg T, pretty func(io.Writer, T) error) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, msg)
	case FormatSilent:
		return nil
	case FormatPretty:
		return pretty(w, msg)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// Convenience wrapper for the common case: reads format and writer from cmd.
func CmdOutput[T proto.Message](cmd *cli.Command, msg T, pretty func(io.Writer, T) error) error {
	format := getFormat(cmd)
	w, err := getWriter(cmd)
	if err != nil {
		return fmt.Errorf("failed to open output: %w", err)
	}
	if w != os.Stdout {
		defer w.Close()
	}
	return Output(format, w, msg, pretty)
}

var jsonOpts = protojson.MarshalOptions{
	UseProtoNames: true,
	Indent:        "  ",
}

func writeJSON(w io.Writer, msg proto.Message) error {
	data, err := jsonOpts.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}
