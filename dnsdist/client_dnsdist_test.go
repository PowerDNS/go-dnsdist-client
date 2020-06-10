package dnsdist

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func randomAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", 30000+rand.Intn(20000))

}

const secret = "FUih6Cj9mNbrHWavhyhUrUhsuWK0pXSXc/d10WeV3h4="

func TestConn_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test against dnsdist in short mode")
	}

	sockAddr := randomAddress()

	conf, err := ioutil.TempFile("", "dnsdist.conf-")
	if err != nil {
		t.Fatal(err)
	}
	defer conf.Close()

	var buf bytes.Buffer
	_, _ = buf.WriteString(fmt.Sprintf("controlSocket(%q)\n", sockAddr))
	_, _ = buf.WriteString(fmt.Sprintf("setKey(%q)\n", secret))
	if _, err := conf.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}

	dnsdistCommand := os.Getenv("TEST_DNSDIST_COMMAND")
	if dnsdistCommand == "" {
		dnsdistCommand = "dnsdist"
	}
	args := []string{
		"--local=" + randomAddress(),
		"--config=" + conf.Name(),
		"--supervised",
		"--disable-syslog",
	}
	t.Logf("dnsdistCommand = %q", dnsdistCommand)
	t.Logf("args = %v", args)

	var outb, errb bytes.Buffer
	cmd := exec.Command(dnsdistCommand, args...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Start(); err != nil {
		t.Fatalf("Could not start command: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()
	time.Sleep(time.Second)

	// Not entirely safe during command execution
	t.Log("stdout:\n", outb.String())
	t.Log("stderr:\n", errb.String())

	// Actual test
	conn, err := Dial(sockAddr, secret)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Try a command
	resp, err := conn.Command("showVersion()")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("showVersion -> %s", resp)
	if !strings.HasPrefix(resp, "dnsdist") {
		t.Errorf("Unexpected version string returned: %q", resp)
	}

	// Lua errors do not result in a Go error
	resp, err = conn.Command("error('some error')")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("error -> %s", resp)
	if !strings.Contains(resp, "some error") {
		t.Errorf("Unexpected error string returned: %q", resp)
	}
}
