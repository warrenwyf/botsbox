package xlog

import (
	"testing"

	"../runtime"
)

func Test_Log(t *testing.T) {
	runtime.LogDir = "/tmp"
	t.Log("Log directory:", runtime.GetAbsLogDir())

	Outf("Outf(%s)\n", "test")
	Outln("Outln(test)")

	Errf("Errf(%s)\n", "test")
	Errln("Errln(test)")

	FlushAll()
	CloseAll()
}
