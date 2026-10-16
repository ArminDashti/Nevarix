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
	"status":               {0, cmdStatus},
	"service status":       {0, cmdServiceStatus},
	"service stop":         {0, cmdServiceStop},
	"service restart":      {0, cmdServiceRestart},
	"service start":        {0, cmdServiceStart},
	"help":                 {0, cmdHelp},
	"version":              {0, cmdVersion},
	"update":               {0, cmdUpdate},
	"install":              {0, cmdInstall},
	"uninstall":            {0, cmdUninstall},
	"rebuild":              {0, cmdRebuild},
	"core address set":     {1, cmdAddressSet},
	"core status":          {0, cmdCoreStatus},
	"core reconnect":       {0, cmdCoreReconnect},
	"core disconnect":      {1, cmdCoreDisconnect},
	"core connect":         {1, cmdCoreConnect},
	"core address list":    {0, cmdAddressList},
	"core address delete":  {1, cmdAddressDelete},
	"core address update":  {1, cmdAddressUpdate},
	"core scan":            {0, cmdCoreScan},
	"core secret-key show": {0, cmdSecretKeyShow},
	"core stats":           {0, cmdCoreStats},
	"app status":           {0, agentApp.Status},
	"app version":          {0, agentApp.Version},
	"app run":              {-1, agentApp.Run},
	"app stop":             {0, agentApp.Stop},
	"app port":             {-1, agentApp.Port},
	"app update":           {0, agentApp.Update},
	"app rebuild":          {-1, agentApp.Rebuild},
	"app uninstall":        {-1, agentApp.Uninstall},
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
