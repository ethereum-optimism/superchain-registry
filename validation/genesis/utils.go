package genesis

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func countImmutables(irs map[string][]ImmutableReference) int {
	count := 0
	for range irs {
		count++
	}
	return count
}

func maskBytecode(b []byte, immutableReferences map[string][]ImmutableReference) error {
	for _, v := range immutableReferences {
		for _, r := range v {
			for i := r.Start; i < r.Start+r.Length; i++ {
				if i >= len(b) {
					return errors.New("immutable references extend beyond bytecode")
				}
				b[i] = 0
			}
		}
	}
	return nil
}

func executeCommandInDir(t *testing.T, dir string, cmd *exec.Cmd) {
	t.Logf("executing %s", cmd.String())
	cmd.Dir = dir
	var outErr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		// error case : status code of command is different from 0
		fmt.Println(outErr.String())
		t.Fatal(err)
	}
}
