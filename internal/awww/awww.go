package awww

import (
	"os/exec"
	"strings"
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
	args := []string{"img"}

	args = append(args, imagePath)

	if a.flags != "" {
		args = append(args, strings.Fields(a.flags)...)
	}

	cmd := exec.Command("awww", args...)
	return cmd.Run()
}
