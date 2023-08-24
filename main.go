package main

import (
	"debug/buildinfo"
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func main() {
	if err := cmd().Execute(); err != nil {
		var exitCode *exitCodeError
		if errors.As(err, &exitCode) {
			os.Exit(exitCode.code)
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(1)
	}
}

// exitCodeError wraps an error with a given exit code.
type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string {
	return e.err.Error()
}

func (e *exitCodeError) Unwrap() error {
	return e.err
}

const long = `gobincache determines whether a Go binary is up-to-date relative to its module
in your go.mod.

It assumes the use of a "tools.go" approach to versioning binaries in your
project.

The command will return an exit code of 0 when the binary currently installed
is up-to-date. It will return an exit code of 2 when it is either not present
or requires updating via "go install".

Any other error will cause this command to exit with a code of 1 (e.g. failing
to parse to go.mod file).
`

// cmd builds the root *cobra.Command hierarchy.
func cmd() *cobra.Command {
	c := &cobra.Command{
		Use:           "gobincache [path to Go binary]",
		Short:         "Determines whether a Go binary requires updating, relative to it's version in the go.mod.",
		Long:          long,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			needsInstall, err := requiresInstall(args[0])
			if err != nil {
				return err
			}
			if needsInstall {
				return &exitCodeError{code: 2, err: fmt.Errorf("binary requires install")}
			}

			return nil
		},
	}

	return c
}

func requiresInstall(binPath string) (bool, error) {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return false, fmt.Errorf("reading modfile: %w", err)
	}

	gomod, err := modfile.Parse("", b, nil)
	if err != nil {
		return false, fmt.Errorf("parsing modfile: %w", err)
	}

	bin, err := moduleFromBinary(binPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}

		return false, fmt.Errorf("reading binary buildinfo (%s): %w", binPath, err)
	}

	mod := versionFromGoMod(gomod, bin)
	if mod == nil {
		return false, fmt.Errorf("module (%s) not found in modfile", bin.Path)
	}

	return bin.Version != mod.Version, nil
}

func moduleFromBinary(path string) (debug.Module, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return debug.Module{}, err
	}

	return info.Main, nil
}

func versionFromGoMod(gomod *modfile.File, binaryModule debug.Module) *module.Version {
	for _, required := range gomod.Require {
		if required.Mod.Path == binaryModule.Path {
			return &required.Mod
		}
	}

	return nil
}
