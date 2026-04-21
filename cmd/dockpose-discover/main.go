// Command dockpose-discover prints discovered Docker Compose stack candidates.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cwklurks/dockpose/internal/config"
	"github.com/cwklurks/dockpose/internal/stack"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dockpose-discover — helper CLI for compose stack discovery\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("dockpose-discover %s\n", version)
		return
	}

	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	reg, err := stack.LoadFromPaths(cfg.ResolvedScanPaths(), cfg.ScanDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover stacks: %v\n", err)
		os.Exit(1)
	}

	if reg.Count() == 0 {
		fmt.Println("No stacks found.")
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PATH\tNAME\tSERVICES\tPROFILES")
	for _, s := range reg.Stacks {
		svcs := make([]string, 0, len(s.Services))
		for _, sv := range s.Services {
			svcs = append(svcs, sv.Name)
		}
		profiles := "-"
		if len(s.Profiles) > 0 {
			profiles = strings.Join(s.Profiles, ",")
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Path, s.Name, strings.Join(svcs, ","), profiles)
	}
	_ = tw.Flush()
}
