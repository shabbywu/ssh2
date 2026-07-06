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

func NewConsole(opts ...expect.ConsoleOpt) (*Console, error) {
	cp, err := expect.NewConsole(opts...)
	return &Console{Console: cp}, err
}

func (c *Console) KillChildren() {
	for _, child := range c.Children {
		if child.Process != nil {
			_ = child.Process.Kill()
		}
	}
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
