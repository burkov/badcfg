package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/burkov/badcfg/badcfglib"
	"github.com/muesli/termenv"
	"golang.design/x/clipboard"
)

type SubCommand string

const (
	SubCommandList SubCommand = "list"
	SubCommandCopy SubCommand = "copy"
	SubCommandHelp SubCommand = "help"
)

var p termenv.Profile

func main() {
	restoreConsole, err := termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
	if err != nil {
		fmt.Println("Error enabling virtual terminal processing:", err)
		os.Exit(1)
	}
	defer restoreConsole()
	p = termenv.ColorProfile()

	subCommand, args := parseSubCommand(os.Args[1:])
	switch subCommand {
	case SubCommandHelp:

	case SubCommandList:
		list(args[0])
	case SubCommandCopy:
		copy(args[0])
	default:
		printError("invalid subcommand '" + string(subCommand) + "'\n")
		printUsage()
		os.Exit(1)
	}
}

func list(pattern string) {
	cfg := mustReadConfig()

	for value := range cfg.Values() {
		if strings.Contains(value.Key, pattern) {
			ppKey := termenv.String(value.Key).Foreground(p.Color("2"))
			shouldMask := value.Type == badcfglib.KdbxConfig
			ppValue := termenv.String(truncate(value.Value, 50, shouldMask)).Faint()
			fmt.Printf("%s = %s\n", ppKey, ppValue)
		}
	}
}

func copy(key string) {
	cfg := mustReadConfig()
	result := make([]badcfglib.ConfigFileEntry, 0)
	for value := range cfg.Values() {
		if strings.Contains(value.Key, key) {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		printError("no value found for key '" + key + "'")
		os.Exit(1)
	}
	if len(result) > 1 {
		printError("more than one value found for key '" + key + "'\n")
		for _, value := range result {
			ppKey := termenv.String(value.Key).Foreground(p.Color("2"))
			ppValue := termenv.String(truncate(value.Value, 50, value.Type == badcfglib.KdbxConfig)).Faint()
			fmt.Printf("  %s = %s\n", ppKey, ppValue)
		}
		fmt.Println()
		os.Exit(1)
	}
	if err := clipboard.Init(); err != nil {
		printError("failed to initialize clipboard: " + err.Error())
		os.Exit(1)
	}
	clipboard.Write(clipboard.FmtText, []byte(result[0].Value))
	ppKey := termenv.String(result[0].Key).Foreground(p.Color("2"))
	ppValue := termenv.String(truncate(result[0].Value, 50, result[0].Type == badcfglib.KdbxConfig)).Faint()
	fmt.Printf("Key %v (%v) copied to clipboard\n", ppKey, ppValue)
}

func mustReadConfig() *badcfglib.BadConfig {
	cfg, err := badcfglib.ReadConfig()
	if err != nil {
		printError("failed to read config: " + err.Error())
		os.Exit(1)
	}
	return cfg
}

func printUsage() {
	tplEngine := template.New("tpl").Funcs(termenv.TemplateFuncs(p))

	tpl, err := tplEngine.Parse(`{{ Color "5" "badcfg" }} is a dead simple tool to list and copy BAD config file values

{{ Bold "Usage:" }}
  
  {{ Color "5" "list" }} {{ Faint "[key]" }}  List all keys in the config file matching the given key
  {{ Color "5" "copy" }} {{ Faint "<key>" }}  Copy a value from the config file matching the given key (exact match)

{{ Bold "Examples:" }}

  badcfg {{ Color "5" "list" }}                                   {{ Faint (Color "2" "# List all keys in the config file") }}
  badcfg {{ Color "5" "list" }} {{ Faint "jetprofile.datasource.dev1" }}        {{ Faint (Color "2" "# List all keys matching jetprofile.datasource.dev1.*") }}
  badcfg {{ Color "5" "copy" }} {{ Faint "jetprofile.datasource.dev1.host" }}   {{ Faint (Color "2" "# Copy the value of jetprofile.datasource.dev1.host") }}
  badcfg {{ Color "5" "copy" }} {{ Faint "jetprofile" }}                        {{ Faint (Color "1" "# This will result in an error, because more than one key matches") }}

`)
	if err != nil {
		printError("failed to parse template: " + err.Error())
		os.Exit(1)
	}

	tpl.Execute(os.Stdout, nil)
}

func parseSubCommand(args []string) (SubCommand, []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}
	if args[0] == string(SubCommandList) {
		key := ""
		if len(args) > 1 {
			key = args[1]
		}
		return SubCommandList, []string{key}
	}
	if args[0] == string(SubCommandHelp) {
		return SubCommandHelp, nil
	}
	if args[0] == string(SubCommandCopy) {
		if len(args) < 2 {
			printError("key is required\n")
			printUsage()
			os.Exit(0)
		}
		return SubCommandCopy, args[1:]
	}
	printError("invalid subcommand '" + args[0] + "'\n")

	os.Exit(0)
	return "", nil
}

func truncate(s string, max int, mask bool) string {
	if mask {
		return truncateAndMask(s, max)
	}
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func truncateAndMask(s string, max int) string {
	truncated := s
	if len(s) > max {
		truncated = s[:max]
	}
	if len(s) < 4 {
		return strings.Repeat("*", len(s))
	}
	return truncated[:2] + strings.Repeat("*", len(truncated)-4) + truncated[len(truncated)-2:]
}

func printError(error string) {
	fmt.Printf("%s: %s\n", termenv.String("Error").Foreground(p.Color("1")), error)
}
