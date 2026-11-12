//-----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
//-----------------------------------------------------------------------------

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxreader"
)

const (
	defaultWidth  = 120
	defaultIndent = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sxpp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	width := fs.Int("width", defaultWidth, "maximum line width")
	indent := fs.Int("indent", defaultIndent, "indentation of bodies and long heads")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: sxpp [flags] [file ...]\n\n"+
			"Pretty-prints symbolic expressions. Reads standard input if no file is given.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *width <= 0 || *indent <= 0 {
		fmt.Fprintln(stderr, "sxpp: -width and -indent must be greater than zero")
		return 2
	}

	opts := sxbuiltins.Options{Width: *width, Indent: *indent}
	out := bufio.NewWriter(stdout)
	defer out.Flush()

	if fs.NArg() == 0 {
		if err := printAll(out, stdin, opts); err != nil {
			fmt.Fprintf(stderr, "sxpp: %v\n", err)
			return 1
		}
		return 0
	}

	status := 0
	for _, name := range fs.Args() {
		if err := printFile(out, name, opts); err != nil {
			fmt.Fprintf(stderr, "sxpp: %s: %v\n", name, err)
			status = 1
		}
	}
	return status
}

func printFile(out io.Writer, name string, opts sxbuiltins.Options) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return printAll(out, f, opts)
}

// printAll prints every expression found in r, each one terminated by a newline.
func printAll(out io.Writer, r io.Reader, opts sxbuiltins.Options) error {
	rd := sxreader.MakeReader(r)
	for {
		obj, err := rd.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if _, err = sxbuiltins.PrintWith(out, obj, opts); err != nil {
			return err
		}
	}
}
