package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blevz/gotty/backend/localcommand"
	"github.com/blevz/gotty/server"
	"github.com/blevz/gotty/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootcmd := &cobra.Command{}
	rootcmd.AddCommand(newLocalReplayCommand())
	rootcmd.AddCommand(newGottyCommand())
	if err := rootcmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func newLocalReplayCommand() *cobra.Command {
	return &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("hi")
		},
		Use: "replay",
	}
}

func newGottyCommand() *cobra.Command {
	var writeFlag bool
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			appOptions := &server.Options{}
			if err := utils.ApplyDefaultValues(appOptions); err != nil {
				return err
			}
			backendOptions := &localcommand.Options{}
			if err := utils.ApplyDefaultValues(backendOptions); err != nil {
				return err
			}
			hostname, err := os.Hostname()
			if err != nil {
				return err
			}
			appOptions.TitleVariables = map[string]interface{}{
				"command":  args[0],
				"argv":     args[1:],
				"hostname": hostname,
			}
			appOptions.PermitWrite = writeFlag
			factory, err := localcommand.NewFactory(args[0], args[1:], backendOptions)
			if err != nil {
				return err
			}
			srv, err := server.New(factory, appOptions)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			gCtx, gCancel := context.WithCancel(context.Background())

			log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))
			errs := make(chan error, 1)
			go func() {
				errs <- srv.Run(ctx, server.WithGracefullContext(gCtx))
			}()
			err = waitSignals(errs, cancel, gCancel)

			if err != nil && err != context.Canceled {
				fmt.Printf("Error: %s\n", err)
				return err
			}
			return nil
		},
		Use: "gotty",
	}
	cmd.PersistentFlags().BoolVarP(&writeFlag, "write", "w", false, "Enable write access")
	return cmd
}

func waitSignals(errs chan error, cancel context.CancelFunc, gracefullCancel context.CancelFunc) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-errs:
		return err

	case s := <-sigChan:
		switch s {
		case syscall.SIGINT:
			gracefullCancel()
			fmt.Println("C-C to force close")
			select {
			case err := <-errs:
				return err
			case <-sigChan:
				fmt.Println("Force closing...")
				cancel()
				return <-errs
			}
		default:
			cancel()
			return <-errs
		}
	}
}
