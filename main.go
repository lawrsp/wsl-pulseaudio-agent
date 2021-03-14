package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	si "github.com/allan-simon/go-singleinstance"

	"github.com/getlantern/systray"
)

const (
	title   = "pulseaudio-agent-gui"
	tooltip = "Helper to interface with pulseaudio.exe service from WSL"
)

var version string
var gitHash string

//go:embed pulseaudio.ico
var icon []byte

var (
	// Program arguments.
	debug bool
	help  bool
	usage string
	cli   = flag.NewFlagSet(title, flag.ContinueOnError)
)

func onReady() {

	systray.SetIcon(icon)
	systray.SetTitle(title)
	systray.SetTooltip(tooltip)

	help := systray.AddMenuItem("Message", "Shows application messages")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Exit", "Exits application")

	go func() {
		for {
			select {
			case <-help.ClickedCh:
				cli.Usage()
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Print("Exiting systray")
}

func run() (err error) {

	go func() {
		serve()
		// If for some reason process breaks - exit
		log.Printf("Quiting - pulseaudio service ended")
		systray.Quit()
	}()

	systray.Run(onReady, onExit)
	return nil
}

func main() {

	// util.NewLogWriter(title, 0, false)

	// Prepare help and parse arguments
	cli.BoolVar(&help, "help", false, "Show help")
	cli.BoolVar(&debug, "debug", false, "Enable verbose debug logging")

	// Build usage string
	var buf strings.Builder
	cli.SetOutput(&buf)
	fmt.Fprintf(&buf, "\n%s\n\nVersion:\n\t%s (%s)\n\t%s\n\n", tooltip, version, runtime.Version(), gitHash)
	fmt.Fprintf(&buf, "Usage:\n\t%s [options]\n\nOptions:\n\n", title)
	cli.PrintDefaults()
	usage = buf.String()

	// do not show usage while parsing arguments
	cli.Usage = func() {}
	if err := cli.Parse(os.Args[1:]); err != nil {
		// util.ShowOKMessage(util.MsgError, title, err.Error())
		os.Exit(1)
	}
	cli.Usage = func() {
		// text := usage
		ShowOkMessageBox(title, usage)
	}

	if help {
		cli.Usage()
		os.Exit(0)
	}

	// util.NewLogWriter(title, 0, debug)

	// Only allow single instance to run
	lockName := filepath.Join(os.TempDir(), title+".lock")
	inst, err := si.CreateLockFile(lockName)
	if err != nil {
		log.Print("Application already running")
		os.Exit(0)
	}
	defer func() {
		// Not necessary at all
		inst.Close()
		os.Remove(lockName)
	}()

	// enter main processing loop
	if err := run(); err != nil {
		// util.ShowOKMessage(util.MsgError, title, err.Error())
	}
}
