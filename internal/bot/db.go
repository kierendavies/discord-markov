package bot

import (
	"encoding/binary"
	"strings"

	"github.com/dgraph-io/badger/v2"
)

var oneB = uint64ToBytes(1)

func uint64ToBytes(i uint64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], i)
	return b[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func (b *Bot) incrementCounts(keys []string) error {
	txn := b.db.NewTransaction(true)
	defer txn.Discard()

	for _, key := range keys {
		keyB := []byte(key)
		newVal := oneB

		item, err := txn.Get(keyB)
		if err == badger.ErrKeyNotFound {
			// Leave newVal = oneB.
		} else if err != nil {
			return err
		} else {
			err = item.Value(func(v []byte) error {
				newVal = uint64ToBytes(bytesToUint64(v) + 1)
				return nil
			})
			if err != nil {
				return err
			}
		}

		err = txn.Set(keyB, newVal)
		if err == badger.ErrTxnTooBig {
			err = txn.Commit()
			if err != nil {
				return err
			}
			txn = b.db.NewTransaction(true)
		} else if err != nil {
			return err
		}
	}
	txn.Commit()

	return nil
}

func (b *Bot) getCounts(prefix string) (map[string]uint64, error) {
	prefixB := []byte(prefix)
	counts := make(map[string]uint64)

	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefixB); it.ValidForPrefix(prefixB); it.Next() {
			item := it.Item()
			key := string(item.Key())
			tokens := strings.Split(key, tokenSeparator)
			lastToken := tokens[len(tokens)-1]

			// Only consider keys which consist of the prefix plus one extra token.
			if len(prefix)+len(lastToken)+1 != len(key) {
				continue
			}

			err := item.Value(func(v []byte) error {
				counts[lastToken] = bytesToUint64(v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return counts, nil
}
