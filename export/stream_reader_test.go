package export

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamReader_CopyUsesWriteTo(t *testing.T) {
	calls := 0
	r := NewStreamReader(func(w io.Writer) error {
		calls++
		_, err := io.WriteString(w, "hello\n")
		return err
	})
	var dst bytes.Buffer
	// wrap dst so that io.Copy cannot use ReaderFrom on the destination
	n, err := io.Copy(struct{ io.Writer }{&dst}, r)
	if err != nil {
		t.Fatal(err)
	}
	if dst.String() != "hello\n" || n != 6 {
		t.Fatalf("got %q (%d bytes)", dst.String(), n)
	}
	if calls != 1 {
		t.Fatalf("write called %d times", calls)
	}
}

func TestStreamReader_Read(t *testing.T) {
	r := NewStreamReader(func(w io.Writer) error {
		_, err := io.WriteString(w, "abcdef")
		return err
	})
	small := make([]byte, 4)
	n, err := r.Read(small)
	if err != nil || string(small[:n]) != "abcd" {
		t.Fatalf("first read: %q %v", small[:n], err)
	}
	// a copy after a partial read returns the rest
	rest, err := io.ReadAll(r)
	if err != nil || string(rest) != "ef" {
		t.Fatalf("rest: %q %v", rest, err)
	}
}

func TestStreamReader_Errors(t *testing.T) {
	wantErr := errors.New("boom")
	failing := func(w io.Writer) error {
		_, _ = io.WriteString(w, "partial")
		return wantErr
	}
	if _, err := io.Copy(io.Discard, NewStreamReader(failing)); !errors.Is(err, wantErr) {
		t.Fatalf("copy: expected %v, got %v", wantErr, err)
	}
	if _, err := io.ReadAll(NewStreamReader(failing)); !errors.Is(err, wantErr) {
		t.Fatalf("read: expected %v, got %v", wantErr, err)
	}
}

func TestWrite_RemovesFileOnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pps")
	err := Write(path, NewStreamReader(func(w io.Writer) error {
		_, _ = io.WriteString(w, "partial")
		return errors.New("boom")
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("expected incomplete file to be removed, stat error: %v", statErr)
	}
}

func TestWrite_WritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pps")
	if err := Write(path, NewStreamReader(func(w io.Writer) error {
		_, err := io.WriteString(w, "{}\n")
		return err
	})); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "{}\n" {
		t.Fatalf("got %q %v", got, err)
	}
}
