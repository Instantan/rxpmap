package rxpmap

import (
	"context"

	"github.com/Instantan/match"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/pb"
)

type Instance interface {
	Query(query string) (map[string][]byte, error)
	Get(key string) ([]byte, bool)
	Listen(ctx context.Context, query string, c chan map[string][]byte) error
	Write(key string, value []byte) error
	WriteBatch(batch map[string][]byte) error
}

type rxpmap struct {
	db *badger.DB
}

func NewPersistent(name string) (Instance, error) {
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
	q, err := match.Compile(query)
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{}
	err = m.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek([]byte("")); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if q.Matches(string(key)) {
				b := make([]byte, item.ValueSize())
				item.ValueCopy(b)
				data[string(key)] = b
			}
		}
		return nil
	})
	return data, err
}

func (m *rxpmap) Listen(ctx context.Context, query string, c chan map[string][]byte) error {
	q, err := match.Compile(query)
	if err != nil {
		return err
	}
	return m.db.Subscribe(ctx, func(kv *badger.KVList) error {
		m := map[string][]byte{}
		parForeach(kv.Kv, func(item *pb.KV) {
			if q.Matches(string(item.Key)) {
				m[string(item.Key)] = item.Value
			}
		})
		c <- m
		return nil
	}, []pb.Match{
		{
			Prefix: []byte(""),
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
