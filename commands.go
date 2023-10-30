package main

import (
	"encoding/binary"
	"flag"
	"fmt"

	"github.com/kevpar/repl-go"
	"go.etcd.io/bbolt"
)

func allCommands() []repl.Command[*state] {
	return []repl.Command[*state]{
		&lsCommand{},
		&catCommand{},
		&cdCommand{},
	}
}

type lsCommand struct{}

func (c *lsCommand) Name() string        { return "ls" }
func (c *lsCommand) Description() string { return "Display entries in the current bucket." }
func (c *lsCommand) ArgHelp() string     { return "" }

func (c *lsCommand) SetupFlags(fs *flag.FlagSet) {}

func (c *lsCommand) Execute(state *state, fs *flag.FlagSet) error {
	return state.db.View(func(tx *bbolt.Tx) error {
		b, err := getBucketFromPath(tx, state.path)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			if v == nil {
				fmt.Printf("d")
			} else {
				fmt.Printf(" ")
			}
			fmt.Printf(" %s\n", k)
			return nil
		})
		return nil
	})
}

type catCommand struct {
	format *string
}

func (c *catCommand) Name() string        { return "cat" }
func (c *catCommand) Description() string { return "Display the value of an entry." }
func (c *catCommand) ArgHelp() string     { return "<NAME>" }

func (c *catCommand) SetupFlags(fs *flag.FlagSet) {
	c.format = fs.String("format", "s", "Format for output.\ns: string\nh: hex\nv: varint\nuv: uvarint\n")
}

func (c *catCommand) Execute(state *state, fs *flag.FlagSet) error {
	return state.db.View(func(tx *bbolt.Tx) error {
		b, err := getBucketFromPath(tx, state.path)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return nil
		}
		v := b.Get([]byte(fs.Arg(0)))
		switch *c.format {
		case "s":
			fmt.Printf("%s\n", v)
		case "h":
			fmt.Printf("%x\n", v)
		case "v":
			n, _ := binary.Varint(v)
			fmt.Printf("%d\n", n)
		case "uv":
			n, _ := binary.Uvarint(v)
			fmt.Printf("%d\n", n)
		}
		return nil
	})
}

type cdCommand struct {
	root *bool
}

func (c *cdCommand) Name() string        { return "cd" }
func (c *cdCommand) Description() string { return "Change the current bucket." }
func (c *cdCommand) ArgHelp() string     { return "<NAME>..." }

func (c *cdCommand) SetupFlags(fs *flag.FlagSet) {
	c.root = fs.Bool("root", false, "Change directory relative to root instead of current dir.")
}

func (c *cdCommand) Execute(state *state, fs *flag.FlagSet) error {
	if *c.root {
		state.path = nil
	}
	for _, p := range fs.Args() {
		if p == ".." {
			state.path = state.path[:len(state.path)-1]
		} else {
			state.path = append(state.path, p)
		}
	}
	return nil
}
