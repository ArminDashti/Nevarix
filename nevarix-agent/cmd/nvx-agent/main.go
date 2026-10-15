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
	argc int // positional args after the phrase
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
}

// route returns the whitelisted command whose phrase and arg count match args exactly.
func route(args []string) (command, []string, bool) {
	for phrase, cmd := range commands {
		words := strings.Fields(phrase)
		if len(args) != len(words)+cmd.argc || !slices.Equal(args[:len(words)], words) {
			continue
		}
		return cmd, args[len(words):], true
	}
	return command{}, nil, false
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
