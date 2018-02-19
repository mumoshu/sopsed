package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func runInBackground(command string, args ...string) error {
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
			log.Printf("read stdout: %s", scanner.Text())
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
			log.Printf("read stderr: %s", scanner.Text())
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
