package runtime

import (
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var stopSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

var (
	errSignaled = errors.New("signaled")

	cancellationDelaySecondsEnv = "CANCELLATION_DELAY_SECONDS"

	defaultCancellationDelaySeconds = 5
)

// IsSignaled returns true if err returned by Wait indicates that
// the program has received SIGINT or SIGTERM.
func IsSignaled(err error) bool {
	return err == errSignaled
}

// handleSignal runs independent goroutine to cancel an environment.
func handleSignal(env *Environment) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, stopSignals...)

	go func() {
		s := <-ch
		delay := getDelaySecondsFromEnv()
		log.Warn().
			Str("signal", s.String()).
			Int("delay", delay).
			Msg("[runtime] got signal")

		time.Sleep(time.Duration(delay) * time.Second)
		env.Cancel(errSignaled)
	}()
}

func getDelaySecondsFromEnv() int {
	delayStr := os.Getenv(cancellationDelaySecondsEnv)
	if len(delayStr) == 0 {
		return defaultCancellationDelaySeconds
	}

	delay, err := strconv.Atoi(delayStr)
	if err != nil {
		log.Warn().Err(err).
			Str("env", delayStr).
			Int("delay", defaultCancellationDelaySeconds).
			Msg("[runtime] set default cancellation delay seconds")

		return defaultCancellationDelaySeconds
	}

	if delay < 0 {
		log.Warn().Err(err).
			Str("env", delayStr).
			Int("delay", 0).
			Msg("[runtime] round up negative cancellation delay seconds to 0s")
		return 0
	}

	return delay
}

// handleSigPipe discards SIGPIPE if the program is running
// as a systemd service.
//
// Background:
//
// systemd interprets programs exited with SIGPIPE as
// "successfully exited" because its default SuccessExitStatus=
// includes SIGPIPE.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html
//
// Normal Go programs ignore SIGPIPE for file descriptors other than
// stdout(1) and stderr(2).  If SIGPIPE is raised from stdout or stderr,
// Go programs exit with a SIGPIPE signal.
// https://golang.org/pkg/os/signal/#hdr-SIGPIPE
//
// journald is a service tightly integrated in systemd.  Go programs
// running as a systemd service will normally connect their stdout and
// stderr pipes to journald.  Unfortunately though, journald often
// dies and restarts due to bugs, and once it restarts, programs
// whose stdout and stderr were connected to journald will receive
// SIGPIPE at their next writes to stdout or stderr.
//
// Combined these specifications and problems all together, Go programs
// running as systemd services often die with SIGPIPE, but systemd will
// not restart them as they "successfully exited" except when the service
// is configured with SuccessExitStatus= without SIGPIPE or Restart=always.
//
// If we catch SIGPIPE and exits abnormally, systemd would restart the
// program.  However, if we call signal.Notify(c, syscall.SIGPIPE),
// SIGPIPE would be raised not only for stdout and stderr but also for
// other file descriptors.  This means that programs that make network
// connections will get a lot of SIGPIPEs and die.  Of course, this is
// not acceptable.
//
// Therefore, we just catch SIGPIPEs and drop them if the program
// runs as a systemd service.  This way, we can detect journald restarts
// by checking the errors from os.Stdout.Write or os.Stderr.Write.
func handleSigPipe() {
	// signal.Ignore does NOT ignore signals; instead, it just stop
	// relaying signals to the channel.  Instead, we set a nop handler.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPIPE)
}
