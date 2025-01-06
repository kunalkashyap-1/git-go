package main

import (
	"fmt"
	"os"

	"github.com/kunalkashyap-1/git-go/internal"
	"github.com/kunalkashyap-1/git-go/internal/strategies"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: git-go <command> [<args>...]\n")
		os.Exit(1)
	}

	var context internal.CommandContext

	switch command := os.Args[1]; command {
	case "init":
		context.SetStrategy(&strategies.InitStrategy{})
	case "cat-file":
		context.SetStrategy(&strategies.CatFileStrategy{})
	case "hash-object":
		context.SetStrategy(&strategies.HashObjectStrategy{})
	case "ls-tree":
		context.SetStrategy((&strategies.LsTreeStrategy{}))
	case "write-tree":
		context.SetStrategy((&strategies.WriteTreeStrategy{}))
	case "commit-tree":
		context.SetStrategy((&strategies.CommitTreeStrategy{}))
	default:
		fmt.Fprintf(os.Stderr, "unknown command %s\n", command)
		os.Exit(1)
	}

	if err := context.Execute(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		os.Exit(1)
	}
}
