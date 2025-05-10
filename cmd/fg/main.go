package main

import (
	"fmt"
	"os"

	"github.com/douglasedurocha/fg/pkg/app"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if err := executeCommand(cmd, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func executeCommand(cmd string, args []string) error {
	switch cmd {
	case "list":
		return app.List()
	case "available":
		return app.Available()
	case "install":
		if len(args) < 1 {
			return fmt.Errorf("version is required for install command")
		}
		return app.Install(args[0])
	case "update":
		return app.Update()
	case "status":
		return app.Status()
	case "logs":
		pid := ""
		if len(args) > 0 {
			pid = args[0]
		}
		return app.Logs(pid)
	case "stop":
		pid := ""
		if len(args) > 0 {
			pid = args[0]
		}
		return app.Stop(pid)
	case "uninstall":
		if len(args) < 1 {
			return fmt.Errorf("version is required for uninstall command")
		}
		return app.Uninstall(args[0])
	case "start":
		version := ""
		if len(args) > 0 {
			version = args[0]
		}
		return app.Start(version)
	case "config":
		if len(args) < 1 {
			return fmt.Errorf("version is required for config command")
		}
		return app.Config(args[0])
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Usage: fg <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  list                      List all installed versions")
	fmt.Println("  available                 List all available versions with release dates")
	fmt.Println("  install <version>         Install a specific version")
	fmt.Println("  update                    Install the latest version")
	fmt.Println("  status                    Show running instances")
	fmt.Println("  logs [pid]                View logs of an instance")
	fmt.Println("  stop [pid]                Stop an instance")
	fmt.Println("  uninstall <version>       Uninstall a version")
	fmt.Println("  start [version]           Start the application")
	fmt.Println("  config <version>          Show configuration of a version")
} 