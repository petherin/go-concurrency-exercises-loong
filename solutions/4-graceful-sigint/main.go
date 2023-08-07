//////////////////////////////////////////////////////////////////////
//
// Given is a mock process which runs indefinitely and blocks the
// program. Right now the only way to stop the program is to send a
// SIGINT (Ctrl-C). Killing a process like that is not graceful, so we
// want to try to gracefully stop the process first.
//
// Change the program to do the following:
//   1. On SIGINT try to gracefully stop the process using
//          `proc.Stop()`
//   2. If SIGINT is called again, just kill the program (last resort)
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a process
	proc := MockProcess{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	done := make(chan struct{}, 1)
	properlyDone := make(chan struct{}, 1)

	// Run the process (blocking)
	go proc.Run()

	go func() {
		fmt.Println("waiting for sig")
		sig := <-sigs
		fmt.Println("signal received: " + sig.String())
		done <- struct{}{}
	}()

	<-done

	go proc.Stop()

	go func() {
		fmt.Println("waiting for second sig")
		sig := <-sigs
		fmt.Println("second signal received: " + sig.String())
		properlyDone <- struct{}{}
	}()

	<-properlyDone

	os.Exit(1)
}
