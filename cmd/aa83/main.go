package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime/debug"
)

var commands []subCommand

type subCommand struct {
	*flag.FlagSet
	cb func()
}

func main() {

	bi, _ := debug.ReadBuildInfo()
	fmt.Println(bi.Main)
	e, _ := os.Executable()
	n := path.Base(e)
	fmt.Println(n) // TODO: UserConfigDir
	fmt.Println(os.UserConfigDir())

	args := os.Args[1:]
	if len(args) < 1 {
		args = []string{"help"}
	}
	verb := args[0]
	args = args[1:]

	commands = []subCommand{
		doHelp(),
		doServe(),
	}

	for _, try := range []string{verb, "help"} {
		for _, cmd := range commands {
			if try == cmd.Name() {
				cmd.Parse(args)
				cmd.cb()
				return
			}
		}
	}

}

func doHelp() subCommand {
	fs := flag.NewFlagSet("help", flag.ContinueOnError)
	cb := func() {
		verb := fs.Arg(0)
	LIST_COMMANDS:
		if verb == "" {
			fmt.Fprintln(os.Stderr, "Available commands:")
			for _, cmd := range commands {
				fmt.Fprintln(os.Stderr, cmd.Name())
			}
		}

		miss := true
		for _, cmd := range commands {
			if cmd.Name() == "help" || (verb != "" && cmd.Name() != verb) {
				continue
			}
			miss = false
			if verb == "" {
				fmt.Fprintln(os.Stderr)
			}
			cmd.Parse([]string{"-help"})
		}
		if miss {
			verb = ""
			goto LIST_COMMANDS
		}
	}

	return subCommand{
		FlagSet: fs,
		cb:      cb,
	}
}

func doServe() subCommand {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":0", "bind addr")

	return subCommand{
		FlagSet: fs,
		cb:      func() { fmt.Println("serve", *addr) },
	}
}
