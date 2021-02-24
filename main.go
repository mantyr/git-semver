package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mantyr/git-semver/v6/version"
)

var prefix = flag.String("prefix", "", "prefix of version string e.g. v (default: none)")
var format = flag.String("format", "", "format string (e.g.: x.y.z-p+m)")
var excludeHash = flag.Bool("no-hash", false, "exclude commit hash (default: false)")
var excludeMeta = flag.Bool("no-meta", false, "exclude build metadata (default: false)")
var setMeta = flag.String("set-meta", "", "set build metadata (default: none)")
var excludePreRelease = flag.Bool("no-pre", false, "exclude pre-release version (default: false)")
var excludePatch = flag.Bool("no-patch", false, "exclude pre-release version (default: false)")
var excludeMinor = flag.Bool("no-minor", false, "exclude pre-release version (default: false)")
var releaseCandidate = flag.Bool("release-candidate", false, "add release candidate (default: false)")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [opts] [<repo>]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func selectFormat() string {
	if *format != "" {
		return *format
	}
	var format string
	switch {
	case *excludeMinor:
		format = version.NoMinorFormat
	case *excludePatch:
		format = version.NoPatchFormat
	case *excludePreRelease:
		format = version.NoPreFormat
	case *excludeHash, *excludeMeta:
		format = version.NoMetaFormat
	case *releaseCandidate:
		format = version.ReleaseCandidate
	default:
		format = version.FullFormat
	}
	return format
}

func main() {
	flag.Parse()
	repoPath := flag.Arg(0)
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	v, err := version.NewFromRepo(repoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *setMeta != "" {
		v.Meta = *setMeta
	}
	if *prefix != "" {
		v.Prefix = *prefix
	}
	s, err := v.Format(selectFormat())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(s)
}
