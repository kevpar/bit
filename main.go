package main

import (
	"fmt"
	"os"

	"github.com/kevpar/repl-go"
	"go.etcd.io/bbolt"
)

type state struct {
	db   *bbolt.DB
	path []string
}

func main() {
	dbPath := os.Args[1]
	db, err := bbolt.Open(dbPath, 0, &bbolt.Options{
		ReadOnly: true,
		OpenFile: func(s string, i int, fm os.FileMode) (*os.File, error) { return os.OpenFile(s, i&^os.O_CREATE, fm) },
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := repl.Run(
		&state{db: db},
		allCommands(),
		func(state *state) string {
			return fmt.Sprintf("%s", state.path)
		},
	); err != nil {
		panic(err)
	}
}

func getBucketFromPath(tx *bbolt.Tx, path []string) (*bbolt.Bucket, error) {
	b := tx.Cursor().Bucket()
	for _, p := range path {
		b = b.Bucket([]byte(p))
		if b == nil {
			return nil, fmt.Errorf("not found: %s", p)
		}
	}
	return b, nil
}
