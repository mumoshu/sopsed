package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

func runAndCaptureStdout(ctx *Context, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	// See
	var wg sync.WaitGroup
	wg.Add(2)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		return "", fmt.Errorf("failed attaching to stdout of %s: %v", command, err)
	}

	var stdoutBuffer bytes.Buffer
	stdoutReader := bufio.NewReader(stdout)
	go func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			ctx.Debug(fmt.Sprintf("read stdout: %s", scanner.Text()))
			stdoutBuffer.WriteString(scanner.Text())
		}
	}(stdoutReader)

	stderr, err := cmd.StderrPipe()
	if nil != err {
		return "", fmt.Errorf("failed attaching to stderr of %s: %v", command, err)
	}

	var stderrBuffer bytes.Buffer
	stderrReader := bufio.NewReader(stderr)
	go func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			ctx.Debug(fmt.Sprintf("read stderr: %s", scanner.Text()))
			stderrBuffer.WriteString(scanner.Text())
		}
	}(stderrReader)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed starting %s: %s: %v", cmd.Path, err.Error(), stderrBuffer.String())
	}

	// Wait until stdout and stderr gets flushed.
	// Otherwise we lose some outputs due to that cmd.Wait terminates pipes before full consumption
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("failed waiting %s: %s: %s", cmd.Path, err.Error(), stderrBuffer.String())
	}

	return stdoutBuffer.String(), nil
}

func runInBackground(ctx *Context, command string, args ...string) error {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		return fmt.Errorf("failed attaching to stdout of %s: %v", command, err)
	}

	var stdoutBuffer bytes.Buffer
	stdoutReader := bufio.NewReader(stdout)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			ctx.Debug(fmt.Sprintf("read stdout: %s", scanner.Text()))
			stdoutBuffer.WriteString(scanner.Text())
		}
	}(stdoutReader)

	stderr, err := cmd.StderrPipe()
	if nil != err {
		return fmt.Errorf("failed attaching to stderr of %s: %v", command, err)
	}

	var stderrBuffer bytes.Buffer
	stderrReader := bufio.NewReader(stderr)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			ctx.Debug(fmt.Sprintf("read stderr: %s", scanner.Text()))
			stderrBuffer.WriteString(scanner.Text())
		}
	}(stderrReader)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed starting %s: %s: %v", cmd.Path, err.Error(), stderrBuffer.String())
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed waiting %s: %s: %s", cmd.Path, err.Error(), stderrBuffer.String())
	}

	return nil
}

func runInForeground(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); nil != err {
		return fmt.Errorf("failed running %s: %s", cmd.Path, err)
	}
	return nil
}
