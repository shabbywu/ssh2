package console

import (
	"github.com/ActiveState/termtest/expect"
	"os/exec"
)

type Console struct {
	Children []*exec.Cmd
	*expect.Console
}

func NewConsole() (*Console, error) {
	cp, err := expect.NewConsole()
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
