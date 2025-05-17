package store

import (
   "fmt"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/mbbgs/rook/consts"
	"github.com/mbbgs/rook/models"
	"github.com/mbbgs/rook/types"
	"github.com/mbbgs/rook/utils"
)

type Store struct {
    db *badger.DB
}


func NewStore() (*Store, error) {
	dir, err := utils.GetSessionDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, consts.STORE_FILE_PATH)
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}




func (s *Store) CreateUser(user *models.User) error {
	exists, _ := UserExists(s.db, user.Username)
	if exists {
		return errors.New("username in use")
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(user.Username), data)
	})
}

func UserExists(db *badger.DB, username string) (bool, error) {
	var exists bool
	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(username))
		if err == nil {
			exists = true
			return nil
		}
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
	return exists, err
}



func (s *Store) GetUser(username string) (*models.User, error) {
	var user models.User
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(username))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})
	return &user, err
}

func (s *Store) UpdateUser(user *models.User) error {
	exists, err := UserExists(s.db, user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user does not exist")
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(user.Username), data)
	})
}

func (s *Store) DeleteUser(username string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(username))
	})
}

func (s *Store) AddToStore(username string, label types.Label, data types.Data) error {
    key := s.makeKey(username, label)
    value, err := json.Marshal(data)
    if err != nil {
        return err
    }
    return s.db.Update(func(txn *badger.Txn) error {
        return txn.Set([]byte(key), value)
    })
}

func (s *Store) GetByLabel(username string, label types.Label) (types.Data, error) {
    var data types.Data
    key := s.makeKey(username, label)
    err := s.db.View(func(txn *badger.Txn) error {
        item, err := txn.Get([]byte(key))
        if err != nil {
            return err
        }
        return item.Value(func(val []byte) error {
            return json.Unmarshal(val, &data)
        })
    })
    return data, err
}

func (s *Store) RemoveFromStore(username string, label types.Label) error {
    key := s.makeKey(username, label)
    return s.db.Update(func(txn *badger.Txn) error {
        return txn.Delete([]byte(key))
    })
}

func (s *Store) GetAllForUser(username string) (map[string]types.Data, error) {
    result := make(map[string]types.Data)
    prefix := []byte(username + ":")
    err := s.db.View(func(txn *badger.Txn) error {
        it := txn.NewIterator(badger.DefaultIteratorOptions)
        defer it.Close()

        for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
            item := it.Item()
            k := item.Key()
            err := item.Value(func(val []byte) error {
                var d types.Data
                if err := json.Unmarshal(val, &d); err != nil {
                    return err
                }
                label := string(k[len(prefix):])
                result[label] = d
                return nil
            })
            if err != nil {
                return err
            }
        }
        return nil
    })
    return result, err
}

func (s *Store) makeKey(username string, label types.Label) string {
    return username + ":" + string(label)
}



func (s *Store) Get(username, label string) ([]byte, error) {
	var val []byte
	key := fmt.Sprintf("%s::%s", username, label)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			val = append([]byte{}, v...)
			return nil
		})
	})
	return val, err
}


func (s *Store) CountForUser(username string) (int, error) {
	prefix := []byte(username + "::")
	count := 0

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // just for keys
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})

	return count, err
}