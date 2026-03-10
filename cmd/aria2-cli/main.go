package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/leechael/aria2-cli-skills/internal/output"
	"github.com/leechael/aria2-cli-skills/internal/rpc"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[1:]
	format := output.FormatPlain
	var filtered []string
	for _, a := range args {
		switch a {
		case "--json":
			format = output.FormatJSON
		case "--plain":
			format = output.FormatPlain
		default:
			filtered = append(filtered, a)
		}
	}

	if len(filtered) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := filtered[0]
	cmdArgs := filtered[1:]

	cfg := rpc.LoadConfig()
	client := rpc.NewClient(cfg)
	printer := output.NewPrinter(format)

	var err error
	switch cmd {
	case "add":
		err = cmdAdd(client, printer, cmdArgs)
	case "remove", "rm":
		err = cmdRemove(client, printer, cmdArgs)
	case "force-remove":
		err = cmdForceRemove(client, printer, cmdArgs)
	case "pause":
		err = cmdPause(client, printer, cmdArgs)
	case "pause-all":
		err = cmdPauseAll(client, printer, cmdArgs)
	case "unpause", "resume":
		err = cmdUnpause(client, printer, cmdArgs)
	case "unpause-all", "resume-all":
		err = cmdUnpauseAll(client, printer, cmdArgs)
	case "status":
		err = cmdStatus(client, printer, cmdArgs)
	case "list", "ls":
		err = cmdList(client, printer, cmdArgs)
	case "stat":
		err = cmdStat(client, printer, cmdArgs)
	case "version":
		err = cmdVersion(client, printer, cmdArgs)
	case "purge":
		err = cmdPurge(client, printer, cmdArgs)
	case "remove-result":
		err = cmdRemoveResult(client, printer, cmdArgs)
	case "shutdown":
		err = cmdShutdown(client, printer, cmdArgs)
	case "force-shutdown":
		err = cmdForceShutdown(client, printer, cmdArgs)
	case "save-session":
		err = cmdSaveSession(client, printer, cmdArgs)
	case "get-option":
		err = cmdGetOption(client, printer, cmdArgs)
	case "set-option":
		err = cmdSetOption(client, printer, cmdArgs)
	case "get-global-option":
		err = cmdGetGlobalOption(client, printer, cmdArgs)
	case "set-global-option":
		err = cmdSetGlobalOption(client, printer, cmdArgs)
	case "get-files":
		err = cmdGetFiles(client, printer, cmdArgs)
	case "get-uris":
		err = cmdGetUris(client, printer, cmdArgs)
	case "get-peers":
		err = cmdGetPeers(client, printer, cmdArgs)
	case "get-servers":
		err = cmdGetServers(client, printer, cmdArgs)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `aria2-cli %s — aria2 JSON-RPC client

USAGE
  aria2-cli [--json|--plain] <command> [args...]

DOWNLOAD COMMANDS
  add              Add download by URI(s)
  remove, rm       Remove a download
  force-remove     Force remove a download
  pause            Pause a download
  pause-all        Pause all downloads
  unpause, resume  Resume a paused download
  unpause-all      Resume all paused downloads

QUERY COMMANDS
  status           Show download status by GID
  list, ls         List downloads (active, waiting, stopped)
  stat             Show global statistics
  version          Show aria2 version info
  get-files        List files in a download
  get-uris         List URIs for a download
  get-peers        List peers for a BitTorrent download
  get-servers      List servers for a download

OPTION COMMANDS
  get-option       Get options for a download
  set-option       Change options for a download
  get-global-option  Get global options
  set-global-option  Change global options

SESSION COMMANDS
  save-session     Save current session to file
  purge            Purge completed/error/removed results
  remove-result    Remove a specific download result
  shutdown         Gracefully shutdown aria2
  force-shutdown   Force shutdown aria2

GLOBAL FLAGS
  --json           Output as JSON (machine-readable)
  --plain          Output as plain text (default)

ENVIRONMENT
  ARIA2_RPC_HOST   aria2 RPC host (default: localhost)
  ARIA2_RPC_PORT   aria2 RPC port (default: 6800)
  ARIA2_RPC_SECRET RPC secret token
  ARIA2_RPC_SECURE Use HTTPS (true/1)

EXAMPLES
  $ aria2-cli add https://example.com/file.zip
  $ aria2-cli list active
  $ aria2-cli --json status 2089b05ecca3d829
  $ aria2-cli stat

Use "aria2-cli <command> --help" for more information about a command.
`, version)
}

func wantHelp(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

// --- Command Implementations ---

func cmdAdd(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) == 0 {
		fmt.Fprintf(os.Stderr, `Add a new download by URI.

USAGE
  aria2-cli add <uri> [uri...] [key=value...]

Multiple URIs for the same file (mirrors) can be given at once.
Options are passed as key=value pairs.

EXAMPLES
  # Download a file
  $ aria2-cli add https://example.com/archive.tar.gz

  # Download with custom output directory
  $ aria2-cli add https://example.com/file.zip dir=/tmp

  # Multiple mirrors for the same file
  $ aria2-cli add https://mirror1.com/file.iso https://mirror2.com/file.iso

  # Magnet link
  $ aria2-cli add "magnet:?xt=urn:btih:..."

  # Limit connections
  $ aria2-cli add https://example.com/big.bin max-connection-per-server=4
`)
		if len(args) == 0 {
			return fmt.Errorf("missing URI")
		}
		return nil
	}

	var uris []string
	opts := map[string]string{}
	for _, a := range args {
		if strings.Contains(a, "=") && !strings.HasPrefix(a, "http") && !strings.HasPrefix(a, "ftp") && !strings.HasPrefix(a, "magnet:") {
			parts := strings.SplitN(a, "=", 2)
			opts[parts[0]] = parts[1]
		} else {
			uris = append(uris, a)
		}
	}

	params := []any{uris}
	if len(opts) > 0 {
		params = append(params, opts)
	}

	result, err := c.Call("aria2.addUri", params...)
	if err != nil {
		return err
	}

	var gid string
	if err := json.Unmarshal(result, &gid); err != nil {
		return err
	}

	p.Hint("download added")
	return p.Print(gid)
}

func cmdRemove(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Remove a download.

USAGE
  aria2-cli remove <gid>

Removes the download identified by GID. This does not remove downloaded files.

EXAMPLES
  $ aria2-cli remove 2089b05ecca3d829
  $ aria2-cli rm 2089b05ecca3d829
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.remove", args[0])
	if err != nil {
		return err
	}
	var gid string
	json.Unmarshal(result, &gid)
	p.Hint("download removed")
	return p.Print(gid)
}

func cmdForceRemove(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Force remove a download.

USAGE
  aria2-cli force-remove <gid>

Like remove, but does not wait for BitTorrent to contact tracker first.

EXAMPLES
  $ aria2-cli force-remove 2089b05ecca3d829
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.forceRemove", args[0])
	if err != nil {
		return err
	}
	var gid string
	json.Unmarshal(result, &gid)
	p.Hint("download force removed")
	return p.Print(gid)
}

func cmdPause(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Pause a download.

USAGE
  aria2-cli pause <gid>

Pauses the download identified by GID. Use "unpause" to resume.

EXAMPLES
  $ aria2-cli pause 2089b05ecca3d829
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.pause", args[0])
	if err != nil {
		return err
	}
	var gid string
	json.Unmarshal(result, &gid)
	p.Hint("download paused")
	return p.Print(gid)
}

func cmdPauseAll(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Pause all active/waiting downloads.

USAGE
  aria2-cli pause-all

EXAMPLES
  $ aria2-cli pause-all
`)
		return nil
	}
	result, err := c.Call("aria2.pauseAll")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("all downloads paused")
	return p.Print(ok)
}

func cmdUnpause(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Resume a paused download.

USAGE
  aria2-cli unpause <gid>

ALIASES
  resume

EXAMPLES
  $ aria2-cli unpause 2089b05ecca3d829
  $ aria2-cli resume 2089b05ecca3d829
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.unpause", args[0])
	if err != nil {
		return err
	}
	var gid string
	json.Unmarshal(result, &gid)
	p.Hint("download resumed")
	return p.Print(gid)
}

func cmdUnpauseAll(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Resume all paused downloads.

USAGE
  aria2-cli unpause-all

ALIASES
  resume-all

EXAMPLES
  $ aria2-cli unpause-all
  $ aria2-cli resume-all
`)
		return nil
	}
	result, err := c.Call("aria2.unpauseAll")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("all downloads resumed")
	return p.Print(ok)
}

func cmdStatus(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) < 1 {
		fmt.Fprintf(os.Stderr, `Show download status.

USAGE
  aria2-cli status <gid>

Displays detailed status for a download: progress, speed, file info, etc.

EXAMPLES
  $ aria2-cli status 2089b05ecca3d829
  $ aria2-cli --json status 2089b05ecca3d829
  $ aria2-cli --json status 2089b05ecca3d829 | jq '.totalLength'
`)
		if len(args) < 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.tellStatus", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdList(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `List downloads.

USAGE
  aria2-cli list [active|waiting|stopped]

Without a filter, lists all downloads. Use a filter to show only
active, waiting, or stopped downloads.

ALIASES
  ls

EXAMPLES
  $ aria2-cli list
  $ aria2-cli ls active
  $ aria2-cli --json list waiting | jq '.[].gid'
  $ aria2-cli list stopped
`)
		return nil
	}

	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}

	var allResults []any

	if filter == "all" || filter == "active" {
		result, err := c.Call("aria2.tellActive")
		if err != nil {
			return fmt.Errorf("tellActive: %w", err)
		}
		var items []any
		json.Unmarshal(result, &items)
		allResults = append(allResults, items...)
	}

	if filter == "all" || filter == "waiting" {
		result, err := c.Call("aria2.tellWaiting", 0, 100)
		if err != nil {
			return fmt.Errorf("tellWaiting: %w", err)
		}
		var items []any
		json.Unmarshal(result, &items)
		allResults = append(allResults, items...)
	}

	if filter == "all" || filter == "stopped" {
		result, err := c.Call("aria2.tellStopped", 0, 100)
		if err != nil {
			return fmt.Errorf("tellStopped: %w", err)
		}
		var items []any
		json.Unmarshal(result, &items)
		allResults = append(allResults, items...)
	}

	if len(allResults) == 0 {
		p.Hint("no downloads")
		return nil
	}

	return p.Print(allResults)
}

func cmdStat(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Show global statistics.

USAGE
  aria2-cli stat

Displays overall download/upload speed, number of active/waiting/stopped downloads.

EXAMPLES
  $ aria2-cli stat
  $ aria2-cli --json stat | jq '.downloadSpeed'
`)
		return nil
	}
	result, err := c.Call("aria2.getGlobalStat")
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdVersion(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Show aria2 version.

USAGE
  aria2-cli version

Displays aria2 version number and enabled features.

EXAMPLES
  $ aria2-cli version
  $ aria2-cli --json version | jq '.version'
`)
		return nil
	}
	result, err := c.Call("aria2.getVersion")
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdPurge(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Purge completed/error/removed download results.

USAGE
  aria2-cli purge

Clears the list of stopped downloads. Does not remove downloaded files.

EXAMPLES
  $ aria2-cli purge
`)
		return nil
	}
	result, err := c.Call("aria2.purgeDownloadResult")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("download results purged")
	return p.Print(ok)
}

func cmdRemoveResult(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Remove a specific download result.

USAGE
  aria2-cli remove-result <gid>

Removes the completed/error/removed download result identified by GID.

EXAMPLES
  $ aria2-cli remove-result 2089b05ecca3d829
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.removeDownloadResult", args[0])
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("download result removed")
	return p.Print(ok)
}

func cmdShutdown(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Gracefully shutdown aria2.

USAGE
  aria2-cli shutdown

Saves session and shuts down aria2 gracefully.

EXAMPLES
  $ aria2-cli shutdown
`)
		return nil
	}
	result, err := c.Call("aria2.shutdown")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("aria2 shutting down")
	return p.Print(ok)
}

func cmdForceShutdown(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Force shutdown aria2.

USAGE
  aria2-cli force-shutdown

Shuts down aria2 immediately without waiting for downloads to finish.

EXAMPLES
  $ aria2-cli force-shutdown
`)
		return nil
	}
	result, err := c.Call("aria2.forceShutdown")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("aria2 force shutting down")
	return p.Print(ok)
}

func cmdSaveSession(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Save current session to file.

USAGE
  aria2-cli save-session

Saves all active/waiting downloads to the session file so they can be
restored on next aria2 startup.

EXAMPLES
  $ aria2-cli save-session
`)
		return nil
	}
	result, err := c.Call("aria2.saveSession")
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("session saved")
	return p.Print(ok)
}

func cmdGetOption(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `Get options for a download.

USAGE
  aria2-cli get-option <gid>

EXAMPLES
  $ aria2-cli get-option 2089b05ecca3d829
  $ aria2-cli --json get-option 2089b05ecca3d829 | jq '.dir'
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.getOption", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdSetOption(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) < 2 {
		fmt.Fprintf(os.Stderr, `Change options for a download.

USAGE
  aria2-cli set-option <gid> <key=value> [key=value...]

Only some options can be changed while downloading. See aria2 docs for details.

EXAMPLES
  $ aria2-cli set-option 2089b05ecca3d829 max-download-limit=100K
  $ aria2-cli set-option 2089b05ecca3d829 max-upload-limit=50K
`)
		if len(args) < 2 || wantHelp(args) {
			return nil
		}
	}
	gid := args[0]
	opts := parseKV(args[1:])
	result, err := c.Call("aria2.changeOption", gid, opts)
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("option updated")
	return p.Print(ok)
}

func cmdGetGlobalOption(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) {
		fmt.Fprintf(os.Stderr, `Get global options.

USAGE
  aria2-cli get-global-option

EXAMPLES
  $ aria2-cli get-global-option
  $ aria2-cli --json get-global-option | jq '.dir'
`)
		return nil
	}
	result, err := c.Call("aria2.getGlobalOption")
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdSetGlobalOption(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) < 1 {
		fmt.Fprintf(os.Stderr, `Change global options.

USAGE
  aria2-cli set-global-option <key=value> [key=value...]

EXAMPLES
  $ aria2-cli set-global-option max-overall-download-limit=1M
  $ aria2-cli set-global-option max-concurrent-downloads=10
`)
		if len(args) < 1 || wantHelp(args) {
			return nil
		}
	}
	opts := parseKV(args)
	result, err := c.Call("aria2.changeGlobalOption", opts)
	if err != nil {
		return err
	}
	var ok string
	json.Unmarshal(result, &ok)
	p.Hint("global option updated")
	return p.Print(ok)
}

func cmdGetFiles(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `List files in a download.

USAGE
  aria2-cli get-files <gid>

Shows file paths, sizes, and selection status for multi-file downloads.

EXAMPLES
  $ aria2-cli get-files 2089b05ecca3d829
  $ aria2-cli --json get-files 2089b05ecca3d829 | jq '.[].path'
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.getFiles", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdGetUris(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `List URIs for a download.

USAGE
  aria2-cli get-uris <gid>

Shows all URIs associated with the download and their status.

EXAMPLES
  $ aria2-cli get-uris 2089b05ecca3d829
  $ aria2-cli --json get-uris 2089b05ecca3d829 | jq '.[].uri'
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.getUris", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdGetPeers(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `List peers for a BitTorrent download.

USAGE
  aria2-cli get-peers <gid>

Shows connected peers, their download/upload speeds, and client info.

EXAMPLES
  $ aria2-cli get-peers 2089b05ecca3d829
  $ aria2-cli --json get-peers 2089b05ecca3d829 | jq '.[].ip'
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.getPeers", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func cmdGetServers(c *rpc.Client, p *output.Printer, args []string) error {
	if wantHelp(args) || len(args) != 1 {
		fmt.Fprintf(os.Stderr, `List servers for a download.

USAGE
  aria2-cli get-servers <gid>

Shows the servers used for the download and current download speed per server.

EXAMPLES
  $ aria2-cli get-servers 2089b05ecca3d829
  $ aria2-cli --json get-servers 2089b05ecca3d829 | jq '.[].servers'
`)
		if len(args) != 1 || wantHelp(args) {
			return nil
		}
	}
	result, err := c.Call("aria2.getServers", args[0])
	if err != nil {
		return err
	}
	var data any
	json.Unmarshal(result, &data)
	return p.Print(data)
}

func parseKV(args []string) map[string]string {
	m := make(map[string]string)
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
