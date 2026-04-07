package awww

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/google/shlex"
)

type Awww struct {
	flags string
}

func New(awwwFlags string) *Awww {
	return &Awww{
		flags: awwwFlags,
	}
}

func (a *Awww) Init() error {
	cmd := exec.Command("awww-daemon")
	return cmd.Start()
}

func (a *Awww) Set(imagePath string) error {
	args := []string{"img", imagePath}

	flags, err := shlex.Split(a.flags)
	if err != nil {
		return err
	}

	args = append(args, flags...)

	cmd := exec.Command("awww", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		return nil
	}

	return fmt.Errorf("exit code: %d\n%s\n%v",
		cmd.ProcessState.ExitCode(),
		stderr.String(),
		err,
	)
}
