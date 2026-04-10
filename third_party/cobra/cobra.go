package cobra

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Command struct {
	Use     string
	Short   string
	Long    string
	Example string
	Run     func(cmd *Command, args []string)
	RunE    func(cmd *Command, args []string) error

	children []*Command
	parent   *Command
	in       io.Reader
	out      io.Writer
	err      io.Writer
	flags    *FlagSet
}

type FlagSet struct {
	set      *flag.FlagSet
	metadata []*Flag
	byName   map[string]*Flag
	order    []string
}

type Flag struct {
	Name         string
	Shorthand    string
	Usage        string
	DefaultValue string
	ValueType    string
	Changed      bool
}

type boolFlag interface { IsBoolFlag() bool }

type trackedValue struct {
	inner   flag.Value
	changed *bool
}

func (v *trackedValue) String() string { return v.inner.String() }
func (v *trackedValue) Set(s string) error {
	*v.changed = true
	return v.inner.Set(s)
}
func (v *trackedValue) IsBoolFlag() bool {
	flag, ok := v.inner.(boolFlag)
	return ok && flag.IsBoolFlag()
}

func (c *Command) Flags() *FlagSet {
	if c.flags != nil {
		return c.flags
	}
	set := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	set.SetOutput(io.Discard)
	c.flags = &FlagSet{set: set, byName: map[string]*Flag{}}
	return c.flags
}

func (f *FlagSet) StringVar(target *string, name, value, usage string) {
	f.add(name, "", usage, value, "string", func(changed *bool) flag.Value {
		return &trackedValue{inner: newStringValue(value, target), changed: changed}
	})
}

func (f *FlagSet) BoolVar(target *bool, name string, value bool, usage string) {
	f.add(name, "", usage, strconv.FormatBool(value), "bool", func(changed *bool) flag.Value {
		return &trackedValue{inner: newBoolValue(value, target), changed: changed}
	})
}

func (f *FlagSet) IntVar(target *int, name string, value int, usage string) {
	f.add(name, "", usage, strconv.Itoa(value), "int", func(changed *bool) flag.Value {
		return &trackedValue{inner: newIntValue(value, target), changed: changed}
	})
}

func (f *FlagSet) Lookup(name string) *Flag {
	return f.byName[name]
}

func (f *FlagSet) Changed(name string) bool {
	flag := f.Lookup(name)
	return flag != nil && flag.Changed
}

func (f *FlagSet) Parse(args []string) ([]string, error) {
	if f == nil {
		return args, nil
	}
	for _, meta := range f.metadata {
		meta.Changed = false
	}
	if err := f.set.Parse(args); err != nil {
		return nil, err
	}
	return f.set.Args(), nil
}

func (f *FlagSet) VisitAll(fn func(*Flag)) {
	if f == nil {
		return
	}
	for _, meta := range f.metadata {
		fn(meta)
	}
}

func (f *FlagSet) add(name, shorthand, usage, def, valueType string, builder func(changed *bool) flag.Value) {
	meta := &Flag{Name: name, Shorthand: shorthand, Usage: usage, DefaultValue: def, ValueType: valueType}
	value := builder(&meta.Changed)
	f.set.Var(value, name, usage)
	f.metadata = append(f.metadata, meta)
	f.byName[name] = meta
}

func newStringValue(value string, target *string) flag.Value {
	*target = value
	return stringValue{target: target}
}

type stringValue struct{ target *string }
func (v stringValue) String() string {
	if v.target == nil {
		return ""
	}
	return *v.target
}
func (v stringValue) Set(s string) error { *v.target = s; return nil }

func newBoolValue(value bool, target *bool) flag.Value {
	*target = value
	return boolValue{target: target}
}

type boolValue struct{ target *bool }
func (v boolValue) String() string {
	if v.target == nil {
		return "false"
	}
	return strconv.FormatBool(*v.target)
}
func (v boolValue) Set(s string) error {
	parsed, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	* v.target = parsed
	return nil
}
func (v boolValue) IsBoolFlag() bool { return true }

func newIntValue(value int, target *int) flag.Value {
	*target = value
	return intValue{target: target}
}

type intValue struct{ target *int }
func (v intValue) String() string {
	if v.target == nil { return "0" }
	return strconv.Itoa(*v.target)
}
func (v intValue) Set(s string) error {
	parsed, err := strconv.Atoi(s)
	if err != nil { return err }
	* v.target = parsed
	return nil
}

func (c *Command) AddCommand(children ...*Command) {
	for _, child := range children {
		if child == nil {
			continue
		}
		child.parent = c
		c.children = append(c.children, child)
	}
}

func (c *Command) Execute() error { return c.execute(os.Args[1:]) }

func (c *Command) execute(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			return c.Help()
		case "help":
			return c.executeHelp(args[1:])
		case "-v", "--version":
			if c.parent == nil {
				if version := c.findChild("version"); version != nil {
					return version.invoke(nil)
				}
			}
		}
		if child := c.findChild(args[0]); child != nil {
			return child.execute(args[1:])
		}
	}

	rest, err := c.parseCommandFlags(args)
	if err != nil {
		return err
	}
	if helpRequested(rest) {
		return c.Help()
	}
	return c.invoke(rest)
}

func (c *Command) invoke(args []string) error {
	if c.RunE != nil {
		return c.RunE(c, args)
	}
	if c.Run != nil {
		c.Run(c, args)
		return nil
	}
	if len(c.children) > 0 {
		return c.Help()
	}
	return nil
}

func (c *Command) parseCommandFlags(args []string) ([]string, error) {
	if c.flags == nil {
		return args, nil
	}
	return c.flags.Parse(args)
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func (c *Command) executeHelp(args []string) error {
	if len(args) == 0 {
		return c.Help()
	}

	child := c.findChild(args[0])
	if child == nil {
		return fmt.Errorf("unknown help topic: %s", args[0])
	}
	if len(args) == 1 {
		return child.Help()
	}
	return child.executeHelp(args[1:])
}

func (c *Command) findChild(name string) *Command {
	for _, child := range c.children {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func (c *Command) Help() error {
	_, err := io.WriteString(c.OutOrStdout(), c.helpText())
	return err
}

func (c *Command) helpText() string {
	var b strings.Builder

	title := strings.TrimSpace(c.Long)
	if title == "" {
		title = strings.TrimSpace(c.Short)
	}
	if title != "" {
		fmt.Fprintf(&b, "%s\n\n", title)
	}

	fmt.Fprintf(&b, "Usage:\n  %s", c.CommandPath())
	if len(c.children) > 0 && c.Run == nil && c.RunE == nil {
		b.WriteString(" <command>")
	}
	if c.flags != nil && len(c.flags.metadata) > 0 {
		b.WriteString(" [flags]")
	}
	b.WriteString("\n")

	if len(c.children) > 0 {
		b.WriteString("\nAvailable Commands:\n")
		children := append([]*Command(nil), c.children...)
		sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
		for _, child := range children {
			fmt.Fprintf(&b, "  %-10s %s\n", child.Name(), child.Short)
		}
	}

	b.WriteString("\nFlags:\n")
	b.WriteString("  -h, --help")
	if c.parent == nil && c.findChild("version") != nil {
		b.WriteString("      Show help for this command\n")
		b.WriteString("  -v, --version   Print version information\n")
	} else {
		b.WriteString("      Show help for this command\n")
	}
	if c.flags != nil {
		for _, meta := range c.flags.metadata {
			fmt.Fprintf(&b, "  --%s", meta.Name)
			padding := 12 - len(meta.Name)
			if padding < 2 { padding = 2 }
			b.WriteString(strings.Repeat(" ", padding))
			b.WriteString(meta.Usage)
			if meta.DefaultValue != "" && meta.DefaultValue != "false" && meta.DefaultValue != "0" {
				fmt.Fprintf(&b, " (default %s)", meta.DefaultValue)
			}
			b.WriteString("\n")
		}
	}

	if example := strings.TrimSpace(c.Example); example != "" {
		fmt.Fprintf(&b, "\nExamples:\n%s\n", example)
	}

	return b.String()
}

func (c *Command) Name() string {
	name := strings.TrimSpace(c.Use)
	if idx := strings.IndexRune(name, ' '); idx >= 0 {
		name = name[:idx]
	}
	return name
}

func (c *Command) CommandPath() string {
	if c.parent == nil {
		return c.Name()
	}
	return c.parent.CommandPath() + " " + c.Name()
}

func (c *Command) InOrStdin() io.Reader {
	if c.in != nil {
		return c.in
	}
	if c.parent != nil {
		return c.parent.InOrStdin()
	}
	return os.Stdin
}

func (c *Command) OutOrStdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	if c.parent != nil {
		return c.parent.OutOrStdout()
	}
	return os.Stdout
}

func (c *Command) ErrOrStderr() io.Writer {
	if c.err != nil {
		return c.err
	}
	if c.parent != nil {
		return c.parent.ErrOrStderr()
	}
	return os.Stderr
}

func (c *Command) SetIn(r io.Reader)  { c.in = r }
func (c *Command) SetOut(w io.Writer) { c.out = w }
func (c *Command) SetErr(w io.Writer) { c.err = w }
