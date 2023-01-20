package rxpmap

import (
	"context"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/pb"
)

type rxpmap struct {
	db *badger.DB
}

func NewPersistent(name string) (*rxpmap, error) {
	return new(name, false)
}

func NewMemory() (*rxpmap, error) {
	return new("", true)
}

func new(name string, inmem bool) (*rxpmap, error) {
	opts := badger.DefaultOptions(name)
	opts.IndexCacheSize = 100 << 20
	opts.InMemory = inmem
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	m := &rxpmap{
		db: db,
	}
	return m, nil
}

func (m *rxpmap) Query(query string) (map[string][]byte, error) {
	q, err := parseQuery(query)
	if err != nil {
		return nil, err
	}
	tx := m.db.NewTransaction(false)
	defer tx.Discard()
	iter := tx.NewIterator(badger.IteratorOptions{
		PrefetchValues: false,
		Prefix:         q.staticPrefix,
	})
	data := map[string][]byte{}
	for iter.Valid() {
		iter.Next()
		item := iter.Item()
		key := item.Key()
		if q.Matches(parseKey(string(key))) {
			b := make([]byte, item.ValueSize())
			item.ValueCopy(b)
			data[string(key)] = b
		}
	}
	return data, nil
}

func (m *rxpmap) Listen(ctx context.Context, query string) error {
	q, err := parseQuery(query)
	if err != nil {
		return err
	}
	return m.db.Subscribe(ctx, func(kv *badger.KVList) error {
		parForeach(kv.Kv, func(item *pb.KV) {
			if q.Matches(parseKey(string(item.Key))) {
				// notify
			}
		})
		return nil
	}, []pb.Match{
		{
			Prefix: q.staticPrefix,
		},
	})
}

func (m *rxpmap) Get(key string) ([]byte, bool) {
	t := m.db.NewTransaction(false)
	defer t.Discard()
	item, err := t.Get([]byte(key))
	if err != nil {
		return []byte{}, false
	}
	b, err := item.ValueCopy(make([]byte, item.ValueSize()))
	if err != nil {
		return []byte{}, false
	}
	return b, true
}

func (m *rxpmap) Write(key string, value []byte) error {
	return m.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (m *rxpmap) WriteBatch(batch map[string][]byte) error {
	return m.db.Update(func(txn *badger.Txn) error {
		for k, v := range batch {
			if err := txn.Set([]byte(k), v); err != nil {
				return err
			}
		}
		return nil
	})
}
