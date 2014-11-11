package main

import (
	"os"
	"os/signal"

	"github.com/aminjam/hipops/command"
	"github.com/mitchellh/cli"
)

// Commands is the mapping of all the available Serf commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	Commands = map[string]cli.CommandFactory{

		"exec": func() (cli.Command, error) {
			return &command.ExecCommand{
				ShutdownCh: makeShutdownCh(),
				Ui:         ui,
			}, nil
		},

		/*
			"api": func() (cli.Command, error) {
				return &command.ApiCommand{
					ShutdownCh: makeShutdownCh(),
					Ui:         ui,
				}, nil
			},
		*/

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}

// makeShutdownCh returns a channel that can be used for shutdown
// notifications for commands. This channel will send a message for every
// interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
