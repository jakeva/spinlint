package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/jakeva/spinlint/pkg/loader"
	"github.com/jakeva/spinlint/pkg/reporter"
	"github.com/jakeva/spinlint/pkg/rules"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "spinlint",
		Short:         "A linter for Spinnaker pipeline definitions",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newValidateCmd())
	return root
}

func newValidateCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "validate <file|glob> [...]",
		Short: "Validate one or more Spinnaker pipeline JSON files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, args, format)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text or json")
	return cmd
}

func runValidate(cmd *cobra.Command, args []string, format string) error {
	if format != "text" && format != "json" {
		return fmt.Errorf("unknown format %q: must be text or json", format)
	}

	rep := reporter.New(cmd.OutOrStdout(), format)
	hasViolations := false

	for _, arg := range args {
		paths, err := loader.Glob(arg)
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: no files matched %q\n", arg)
			continue
		}

		for _, path := range paths {
			pipeline, err := loader.LoadFile(path)
			if err != nil {
				return err
			}

			var violations []rules.Violation
			for _, rule := range rules.All {
				violations = append(violations, rule.Check(pipeline)...)
			}

			rep.Add(path, violations)
			if len(violations) > 0 {
				hasViolations = true
			}
		}
	}

	if err := rep.Flush(); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	if hasViolations {
		return fmt.Errorf("one or more files failed validation")
	}
	return nil
}
