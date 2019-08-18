package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midispy"
	driver "gitlab.com/gomidi/rtmididrv"
)

var (
	device = flag.Int("device", 1, "input device")
	list   = flag.Bool("list", false, "list midi device")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	drv, err := driver.New()
	if err != nil {
		return err
	}
	defer drv.Close()

	if *list {
		listMIDIInDevices(drv)
		return nil
	}

	if err = startSpying(drv); err != nil {
		return err
	}

	sigchan := make(chan os.Signal, 10)
	go signal.Notify(sigchan, os.Interrupt)
	<-sigchan

	return nil
}

func listMIDIInDevices(d mid.Driver) {
	ins, _ := d.Ins()
	fmt.Print("\n--- MIDI input ports ---\n\n")
	for _, port := range ins {
		fmt.Printf("[%d] %#v\n", port.Number(), port.String())
	}
}

func startSpying(d mid.Driver) error {
	inPort, err := mid.OpenIn(d, *device, "")
	if err != nil {
		return err
	}

	rd := mid.NewReader(mid.NoLogger())
	rd.Msg.Each = func(_ *mid.Position, m midi.Message) {
		fmt.Printf("%v\n", m)
	}

	return midispy.Run(inPort, nil, rd)
}
