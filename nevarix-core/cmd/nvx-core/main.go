package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// version is set at build time: go build -ldflags "-X main.version=1.0.0"
var version = "dev"

type handler func(args []string) error

type command struct {
	argc int // positional args after the phrase; -1 = handler parses its own flags
	run  handler
}

// commands is the complete whitelist. Anything not listed here is rejected.
var commands = map[string]command{
	"status":              {0, cmdStatus},
	"service status":      {0, cmdServiceStatus},
	"service start":       {0, cmdServiceStart},
	"service stop":        {0, cmdServiceStop},
	"service restart":     {0, cmdServiceRestart},
	"help":                {0, cmdHelp},
	"version":             {0, cmdVersion},
	"install":             {0, cmdInstall},
	"uninstall":           {0, cmdUninstall},
	"rebuild":             {0, cmdRebuild},
	"rebuild remove-data": {0, cmdRebuildRemoveData},
	"agent list":          {0, cmdAgentList},
	"agent add":           {-1, cmdAgentAdd},
	"agent remove":        {1, cmdAgentRemove},
	"agent enable":        {1, cmdAgentEnable},
	"agent disable":       {1, cmdAgentDisable},
	"agent scan":          {0, cmdAgentScan},
	"secret-key show":     {0, cmdSecretKeyShow},
	"management list":     {0, cmdManagementList},
	"stats":               {0, cmdStats},
	"stats reset":         {0, cmdStatsReset},
}

// route returns the longest whitelisted phrase that prefixes args, plus the remaining args.
func route(args []string) (command, []string, bool) {
	var best command
	var rest []string
	bestLen := 0
	found := false
	for phrase, cmd := range commands {
		words := strings.Fields(phrase)
		if len(words) <= bestLen || len(args) < len(words) || !slices.Equal(args[:len(words)], words) {
			continue
		}
		remaining := args[len(words):]
		if cmd.argc >= 0 && len(remaining) != cmd.argc {
			continue
		}
		best, rest, bestLen, found = cmd, remaining, len(words), true
	}
	return best, rest, found
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		cmdHelp(nil)
		return
	}
	cmd, rest, ok := route(args)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", strings.Join(args, " "))
		cmdHelp(nil)
		os.Exit(2)
	}
	if err := cmd.run(rest); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
