package console

import (
	"github.com/ActiveState/termtest/expect"
	"os"
	"os/exec"
)

type Console struct {
	Children []*exec.Cmd
	*expect.Console
}

func NewConsole() (*Console, error) {
	cp, err := expect.NewConsole(expect.WithStdin(os.Stdin))
	return &Console{Console: cp}, err
}

func (c *Console) Wait() error {
	for _, child := range c.Children {
		if err := child.Wait(); err != nil {
			return err
		}
	}
	return nil
}
