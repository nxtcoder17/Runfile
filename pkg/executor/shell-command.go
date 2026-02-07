package executor

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// NewInteractiveShellCommand creates a Command that runs an exec.Cmd with PTY for full terminal support
func NewInteractiveShellCommand(handler func(context.Context) *exec.Cmd) *command {
	return CommandFunc(func(ctx context.Context) error {
		cmd := handler(ctx)

		// Clear these - pty.Start sets them to TTY but won't override existing values
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil

		ptmx, err := pty.Start(cmd)
		if err != nil {
			return err
		}
		defer ptmx.Close()

		// Handle terminal resize
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		go func() {
			for range sigCh {
				pty.InheritSize(os.Stdin, ptmx)
			}
		}()
		sigCh <- syscall.SIGWINCH // Initial resize
		defer signal.Stop(sigCh)

		// Set stdin to raw mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Copy I/O
		go io.Copy(ptmx, os.Stdin)
		io.Copy(os.Stdout, ptmx)

		return cmd.Wait()
	})
}

// NewShellCommand creates a Command that runs an exec.Cmd with lifecycle management
func NewShellCommand(handler func(context.Context) *exec.Cmd) *command {
	return CommandFunc(func(ctx context.Context) error {
		cmd := handler(ctx)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if err := cmd.Start(); err != nil {
			return err
		}

		pid := cmd.Process.Pid
		done := make(chan error, 1)

		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			syscall.Kill(-pid, syscall.SIGTERM)

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				syscall.Kill(-pid, syscall.SIGKILL)
				<-done
			}
			return ctx.Err()
		}
	})
}
