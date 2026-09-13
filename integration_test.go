package main

import (
	"io"
	"os"
	"svm/assembler"
	"svm/machine"
	"testing"
)

func captureStdout(t *testing.T, action func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = previous
		writer.Close()
		reader.Close()
	}()

	action()
	os.Stdout = previous
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestExample00EndToEnd(t *testing.T) {
	code, err := assembler.Assemble("examples/example00.asm")
	if err != nil {
		t.Fatal(err)
	}

	m := machine.NewMachine()
	if err := m.Load(code); err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, m.Run)
	if output != "777\n" {
		t.Fatalf("Սպասվում էր %q արտածումը, ստացվել է %q", "777\\n", output)
	}
}
