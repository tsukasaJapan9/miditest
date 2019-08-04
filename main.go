package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midispy"
	driver "gitlab.com/gomidi/rtmididrv"
	"gitlab.com/metakeule/config"
)

var (
	cfg      = config.MustNew("midispy", "1.9.1", "spy on the MIDI data that is sent from a device to another.")
	inArg    = cfg.NewInt32("in", "number of the input device", config.Required, config.Shortflag('i'))
	outArg   = cfg.NewInt32("out", "number of the output device", config.Shortflag('o'))
	noLogArg = cfg.NewBool("nolog", "don't log, just connect in and out", config.Shortflag('n'))
	shortArg = cfg.NewBool("short", "log the short way", config.Shortflag('s'))
	listCmd  = cfg.MustCommand("list", "list devices").Relax("in").Relax("out")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() (err error) {

	drv, err := driver.New()

	if err != nil {
		return
	}

	if err = cfg.Run(); err != nil {
		listMIDIInDevices(drv)
		return err
	}

	// closing all ports
	defer drv.Close()

	if cfg.ActiveCommand() == listCmd {
		listMIDIDevices(drv)
		return nil
	}

	err = startSpying(drv, !noLogArg.Get())

	if err != nil {
		return err
	}

	sigchan := make(chan os.Signal, 10)

	// listen for ctrl+c
	go signal.Notify(sigchan, os.Interrupt)

	// interrupt has happend
	<-sigchan
	fmt.Println("\n--interrupted!")

	return nil
}

func listMIDIDevices(d mid.Driver) {
	listMIDIInDevices(d)

	outs, _ := d.Outs()

	fmt.Print("\n--- MIDI output ports ---\n\n")

	for _, port := range outs {
		fmt.Printf("[%d] %#v\n", port.Number(), port.String())
	}

	return
}

func listMIDIInDevices(d mid.Driver) {
	ins, _ := d.Ins()

	fmt.Print("\n--- MIDI input ports ---\n\n")

	for _, port := range ins {
		fmt.Printf("[%d] %#v\n", port.Number(), port.String())
	}
}

func startSpying(d mid.Driver, shouldlog bool) error {

	in := inArg.Get()

	inPort, err := mid.OpenIn(d, int(in), "")
	if err != nil {
		return err
	}

	var outPort mid.Out = nil
	var logfn func(...interface{})

	if outArg.IsSet() {

		out := outArg.Get()
		outPort, err = mid.OpenOut(d, int(out), "")
		if err != nil {
			return err
		}

		fmt.Printf("[%d] %#v\n->\n[%d] %#v\n-----------------------\n",
			inPort.Number(), inPort.String(), outPort.Number(), outPort.String())
		logfn = logger(in, out)
	} else {
		fmt.Printf("[%d] %#v\n-----------------------\n",
			inPort.Number(), inPort.String())
		logfn = logger(in, 0)
	}

	rd := mid.NewReader(mid.NoLogger())
	if shouldlog {
		rd.Msg.Each = func(_ *mid.Position, m midi.Message) {
			logfn(m)
		}
	}

	return midispy.Run(inPort, outPort, rd)
}

func logger(in, out int32) func(...interface{}) {
	if shortArg.Get() {
		return func(v ...interface{}) {
			fmt.Println(v...)
		}
	}
	if outArg.IsSet() {
		l := log.New(os.Stdout, fmt.Sprintf("[%d->%d] ", in, out), log.Lmicroseconds)
		return l.Println
	}

	l := log.New(os.Stdout, fmt.Sprintf("[%d] ", in), log.Lmicroseconds)
	return l.Println
}
