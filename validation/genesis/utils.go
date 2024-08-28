package genesis

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func executeCommandInDir(dir string, cmd *exec.Cmd) error {
	log.Printf("executing %s", cmd.String())
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

func mustExecuteCommandInDir(dir string, cmd *exec.Cmd) {
	err := executeCommandInDir(dir, cmd)
	if err != nil {
		panic(err)
	}
}
