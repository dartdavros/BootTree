package cobra

import (
	"fmt"
	"io"
	"os"
)

type Command struct {
	Use   string
	Short string
	Long  string
	Run   func(cmd *Command, args []string)

	children []*Command
	out      io.Writer
}

func (c *Command) AddCommand(children ...*Command) { c.children = append(c.children, children...) }

func (c *Command) Execute() error { return c.execute(os.Args[1:]) }

func (c *Command) execute(args []string) error {
	if len(args) == 0 {
		if c.Run != nil {
			c.Run(c, nil)
		}
		return nil
	}
	for _, child := range c.children {
		if child.Use == args[0] {
			if child.Run != nil {
				child.Run(child, args[1:])
				return nil
			}
			return child.execute(args[1:])
		}
	}
	return fmt.Errorf("unknown command: %s", args[0])
}

func (c *Command) OutOrStdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}

func (c *Command) SetOut(w io.Writer) { c.out = w }
