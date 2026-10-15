package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/core"
	"github.com/ArminDashti/nevarix-agent/internal/paths"
	"github.com/ArminDashti/nevarix-agent/internal/service"
)

var stdin = bufio.NewReader(os.Stdin)

func cmdHelp(_ []string) error {
	fmt.Println(`Usage: nvx-agent <command>

  status
  service status | stop | restart | start
  help
  version
  install
  update
  uninstall
  rebuild
  core address set <alias:address:port>
  core status
  core reconnect
  core disconnect <alias>
  core connect <alias>
  core address list
  core address delete <alias>
  core address update <alias>
  core scan
  core secret-key show
  core stats`)
	return nil
}

func cmdVersion(_ []string) error {
	fmt.Println(version)
	return nil
}

func cmdStatus(_ []string) error {
	if out, err := service.Status(); err != nil {
		fmt.Println("service: not available:", err)
	} else {
		fmt.Println(out)
	}
	return cmdCoreStatus(nil)
}

func cmdServiceStatus(_ []string) error {
	out, err := service.Status()
	fmt.Println(out)
	return err
}

func cmdServiceStop(_ []string) error    { return service.Stop() }
func cmdServiceStart(_ []string) error   { return service.Start() }
func cmdServiceRestart(_ []string) error { return service.Restart() }

func cmdUpdate(_ []string) error {
	if !confirm("Pull latest source and rebuild the agent? [y/N]: ") {
		return nil
	}
	if err := runTool("git", "pull", "--ff-only"); err != nil {
		return err
	}
	return cmdRebuild(nil)
}

func cmdRebuild(_ []string) error {
	if !confirm("Rebuild agent binaries into bin/? [y/N]: ") {
		return nil
	}
	ldflags := "-X main.version=" + version
	if err := runTool("go", "build", "-o", "bin/nevarix-agent.exe", "./cmd/agent"); err != nil {
		return err
	}
	return runTool("go", "build", "-ldflags", ldflags, "-o", "bin/nvx-agent.exe", "./cmd/nvx-agent")
}

func cmdUninstall(_ []string) error {
	if !confirm("Stop and delete the service and local data? [y/N]: ") {
		return nil
	}
	_ = service.Stop() // ponytail: ignores stop error so uninstall still removes a stopped service
	if err := service.Delete(); err != nil {
		return err
	}
	// Removes the agent data directory (outbox, core state). The binary is left in place: it is the running file.
	return os.RemoveAll(paths.DataDir())
}

func cmdAddressSet(args []string) error {
	alias, addr, err := core.ParseEntry(args[0])
	if err != nil {
		return err
	}
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	if _, exists := st.Addresses[alias]; exists {
		return fmt.Errorf("alias %q already exists; use core address update", alias)
	}
	st.Addresses[alias] = addr
	return core.Save(core.StatePath(), st)
}

func cmdAddressList(_ []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	for alias, a := range st.Addresses {
		mark := " "
		if alias == st.Active {
			mark = "*"
		}
		fmt.Printf("%s %s\t%s:%d\n", mark, alias, a.Address, a.Port)
	}
	return nil
}

func cmdAddressDelete(args []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	if _, ok := st.Addresses[args[0]]; !ok {
		return fmt.Errorf("alias %q not found", args[0])
	}
	delete(st.Addresses, args[0])
	if st.Active == args[0] {
		st.Active = ""
	}
	return core.Save(core.StatePath(), st)
}

func cmdAddressUpdate(args []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	if _, ok := st.Addresses[args[0]]; !ok {
		return fmt.Errorf("alias %q not found", args[0])
	}
	fmt.Printf("New address:port for %s: ", args[0])
	line, err := stdin.ReadString('\n')
	if err != nil {
		return err
	}
	addr, err := core.ParseHostPort(strings.TrimSpace(line))
	if err != nil {
		return err
	}
	st.Addresses[args[0]] = addr
	return core.Save(core.StatePath(), st)
}

func cmdCoreConnect(args []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	addr, ok := st.Addresses[args[0]]
	if !ok {
		return fmt.Errorf("alias %q not found", args[0])
	}
	url, _ := core.State{Active: args[0], Addresses: st.Addresses}.CoreURL()
	if err := coreHealth(url); err != nil {
		return fmt.Errorf("cannot reach %s:%d: %w", addr.Address, addr.Port, err)
	}
	st.Active = args[0]
	if err := core.Save(core.StatePath(), st); err != nil {
		return err
	}
	fmt.Printf("connected to %s (restart the service to apply)\n", args[0])
	return nil
}

func cmdCoreDisconnect(args []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	if st.Active != args[0] {
		return fmt.Errorf("alias %q is not connected", args[0])
	}
	st.Active = ""
	if err := core.Save(core.StatePath(), st); err != nil {
		return err
	}
	fmt.Printf("disconnected from %s (restart the service to apply)\n", args[0])
	return nil
}

func cmdCoreReconnect(_ []string) error {
	url, err := activeCoreURL()
	if err != nil {
		return err
	}
	if err := coreHealth(url); err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}
	fmt.Println("reconnected to", url)
	return nil
}

func cmdCoreStatus(_ []string) error {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return err
	}
	url, ok := st.CoreURL()
	if !ok {
		fmt.Println("core: not connected")
		return nil
	}
	if err := coreHealth(url); err != nil {
		fmt.Printf("core: %s (%s) unreachable: %v\n", st.Active, url, err)
		return nil
	}
	fmt.Printf("core: %s (%s) online\n", st.Active, url)
	return nil
}

func cmdCoreScan(_ []string) error {
	url, err := activeCoreURL()
	if err != nil {
		return err
	}
	ctx, cancel := coreContext()
	defer cancel()
	raw, err := core.Scan(ctx, url)
	fmt.Println(string(raw))
	return err
}

func cmdCoreStats(_ []string) error {
	url, err := activeCoreURL()
	if err != nil {
		return err
	}
	ctx, cancel := coreContext()
	defer cancel()
	raw, err := core.Agents(ctx, url)
	fmt.Println(string(raw))
	return err
}

func cmdSecretKeyShow(_ []string) error {
	key := strings.TrimSpace(os.Getenv("NEVARIX_AGENT_TOKEN"))
	if key == "" {
		return errors.New("NEVARIX_AGENT_TOKEN is not set")
	}
	fmt.Println(key)
	return nil
}

// activeCoreURL returns the connected Core URL or an error telling the user to connect first.
func activeCoreURL() (string, error) {
	st, err := core.Load(core.StatePath())
	if err != nil {
		return "", err
	}
	url, ok := st.CoreURL()
	if !ok {
		return "", errors.New("not connected; run core connect <alias>")
	}
	return url, nil
}

func coreHealth(url string) error {
	ctx, cancel := coreContext()
	defer cancel()
	return core.Health(ctx, url)
}

func coreContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	line, _ := stdin.ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

// runTool runs an external tool in the current directory, streaming its output.
func runTool(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
