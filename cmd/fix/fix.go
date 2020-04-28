package main

import (
	"bytes"
	"flag"
	"log"

	"github.com/dgraph-io/badger/v2"
)

func main() {
	dbPath := flag.String("db", "", "db")
	flag.Parse()

	dbOpts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(dbOpts)
	if err != nil {
		log.Fatal(err)
	}

	marker := []byte("\x02 \x03")

	err = db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if bytes.Contains(k, marker) {
				err := txn.Delete(k)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
