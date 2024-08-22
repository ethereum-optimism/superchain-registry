package genesis

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func executeCommandInDir(t *testing.T, dir string, cmd *exec.Cmd) error {
	t.Logf("executing %s", cmd.String())
	cmd.Dir = dir
	var outErr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		// error case : status code of command is different from 0
		fmt.Println(outErr.String())
	}
	return err
}

func mustExecuteCommandInDir(t *testing.T, dir string, cmd *exec.Cmd) {
	err := executeCommandInDir(t, dir, cmd)
	if err != nil {
		t.Fatal(err)
	}
}
