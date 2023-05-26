package log

import (
	"os"
	"testing"
)

func TestSetLevel(t *testing.T) {
	Info("-------->info")

	SetLevel(ErrorLevel)
	if infoLog.Writer() == os.Stdout || errorLog.Writer() != os.Stdout {
		t.Fatal("fail to set log level")
	}
	Error("------->error")

	SetLevel(Disable)
	if infoLog.Writer() == os.Stdout || errorLog.Writer() == os.Stderr {
		t.Fatal("failed to set log level")
	}
	Error("------->disable")
}
