package gsd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"time"
)

const controllerDrainTimeout = time.Second

var errControllerDrainTimeout = errors.New("controller output drain exceeded its bound")

func runProcessTree(cmd *exec.Cmd) error {
	stdoutDestination := cmd.Stdout
	stderrDestination := cmd.Stderr
	stdoutReader, stdoutWriter, err := controllerOutputPipe(stdoutDestination)
	if err != nil {
		return err
	}
	stderrReader, stderrWriter, err := controllerOutputPipe(stderrDestination)
	if err != nil {
		closeControllerPipe(stdoutReader, stdoutWriter)
		return err
	}
	if stdoutWriter != nil {
		cmd.Stdout = stdoutWriter
	}
	if stderrWriter != nil {
		cmd.Stderr = stderrWriter
	}
	stdoutDone := copyControllerOutput(stdoutReader, stdoutDestination)
	stderrDone := copyControllerOutput(stderrReader, stderrDestination)
	if err := cmd.Start(); err != nil {
		closeControllerPipe(stdoutReader, stdoutWriter)
		closeControllerPipe(stderrReader, stderrWriter)
		return errors.Join(err, <-stdoutDone, <-stderrDone)
	}
	closeErr := errors.Join(closeWriter(stdoutWriter), closeWriter(stderrWriter))
	waitErr := cmd.Wait()
	cleanupErr := cleanupProcessTree(cmd)
	copyErr := waitControllerOutputs(stdoutReader, stdoutDone, stderrReader, stderrDone)
	closeControllerPipe(stdoutReader, nil)
	closeControllerPipe(stderrReader, nil)
	return errors.Join(closeErr, waitErr, cleanupErr, copyErr)
}

func controllerOutputPipe(destination io.Writer) (*os.File, *os.File, error) {
	if destination == nil {
		return nil, nil, nil
	}
	return os.Pipe()
}

func copyControllerOutput(reader *os.File, destination io.Writer) <-chan error {
	done := make(chan error, 1)
	if reader == nil {
		done <- nil
		return done
	}
	go func() {
		_, err := io.Copy(destination, reader)
		done <- err
	}()
	return done
}

func waitControllerOutputs(firstReader *os.File, firstDone <-chan error,
	secondReader *os.File, secondDone <-chan error) error {
	results := make(chan error, 2)
	go func() { results <- waitControllerOutput(firstReader, firstDone) }()
	go func() { results <- waitControllerOutput(secondReader, secondDone) }()
	return errors.Join(<-results, <-results)
}

func waitControllerOutput(reader *os.File, done <-chan error) error {
	select {
	case err := <-done:
		return err
	case <-time.After(controllerDrainTimeout):
		if reader != nil {
			_ = reader.Close()
		}
		return errors.Join(errControllerDrainTimeout, <-done)
	}
}

func closeWriter(writer *os.File) error {
	if writer == nil {
		return nil
	}
	return writer.Close()
}

func closeControllerPipe(reader, writer *os.File) {
	if writer != nil {
		_ = writer.Close()
	}
	if reader != nil {
		_ = reader.Close()
	}
}
