package console

import (
	"github.com/ActiveState/termtest/expect"
	"os/exec"
)

type Console struct {
	Children []*exec.Cmd
	Cleanups []func()
	*expect.Console
}

func NewConsole() (*Console, error) {
	cp, err := expect.NewConsole()
	return &Console{Console: cp}, err
}

func (c *Console) Wait() error {
	defer func() {
		for _, cleanup := range c.Cleanups {
			cleanup()
		}
	}()
	for _, child := range c.Children {
		if err := child.Wait(); err != nil {
			return err
		}
	}
	return nil
}
