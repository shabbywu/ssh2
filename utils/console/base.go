package console

import (
	"github.com/ActiveState/termtest/expect"
	"os"
	"os/exec"
)

type Console struct {
	Children []*exec.Cmd
	*expect.Console
	closed chan interface{}
}

func NewConsole() (*Console, error) {
	cp, err := expect.NewConsole(expect.WithStdin(os.Stdin))
	return &Console{Console: cp, closed: make(chan interface{})}, err
}

func (c *Console) Wait() error {
	for _, child := range c.Children {
		if err := child.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) Close() error {
	close(c.closed)
	return c.Console.Close()
}
