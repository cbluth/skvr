package main

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

var (
	db,
	defaultNamespace,
	defaultKey = openDB()
)

type (
	key       []byte
	namespace []byte
)

func (k key) String() string {
	return string(k)
}

func (n namespace) String() string {
	return string(n)
}

func openDB() (*bolt.DB, namespace, key) {
	skvrDir := getENV("SKVR_DIR")
	if skvrDir == "" {
		log.Fatalln("missing dir")
	}
	if err := os.MkdirAll(skvrDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}
	db, err := bolt.Open(skvrDir+"/db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	ns := namespace(getENV("SKVR_DEFAULT_NAMESPACE"))
	err = db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(ns)
			return err
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
	k := key(getENV("SKVR_INDEX_KEY"))
	return db, ns, k
}

func getBuckets() []string {
	b := []string{}
	err := db.View(
		func(tx *bolt.Tx) error {
			return tx.ForEach(
				func(name []byte, _ *bolt.Bucket) error {
					b = append(b, string(name))
					return nil
				},
			)
		},
	)
	if err != nil {
		log.Fatalln(err)
		return []string{}
	}
	return b
}

func (n namespace) listKeys() ([]string, error) {
	keys := []string{}
	err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(n)
			if b == nil {
				return fmt.Errorf("namespace not found: %s", string(n))
			}
			return b.ForEach(
				func(k, v []byte) error {
					keys = append(keys, string(k))
					return nil
				},
			)
		},
	)
	return keys, err
}

func (n namespace) kvGet(k []byte) ([]byte, error) {
	v := []byte{}
	err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(n)
			if b == nil {
				return fmt.Errorf("namespace not found: %s", string(n))
			}
			value := b.Get(k)
			if value == nil {
				return fmt.Errorf("key not found: %s :: %s", string(n), string(k))
			}
			v = value
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (n namespace) kvPut(k, v []byte) error {
	err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(n)
			if b == nil {
				err := *new(error)
				b, err = tx.CreateBucketIfNotExists(n)
				if err != nil {
					return err
				}
				log.Println("Created Namespace:", string(n))
			}
			if k == nil || v == nil {
				return nil
			}
			return b.Put(k, v)
		},
	)
	if err != nil {
		return err
	}
	go func() {
		err := db.Sync()
		if err != nil {
			log.Println(err)
		}
	}()
	return nil
}

func (n namespace) kvDelete(k []byte) error {
	return db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(n)
			if b == nil {
				return fmt.Errorf("namespace not found: %s", string(n))
			}
			value := b.Get(k)
			if value == nil {
				return fmt.Errorf("key not found: %s :: %s", string(n), string(k))
			}
			return b.Delete(k)
		},
	)
}

func (n namespace) delete() error {
	return db.Update(
		func(tx *bolt.Tx) error {
			return tx.DeleteBucket(n)
		},
	)
}

func (n namespace) exists(k key) bool {
	err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(n)
			if b == nil {
				return fmt.Errorf("namespace not found: %s", n.String())
			}
			if k != nil {
				value := b.Get(k)
				if value == nil {
					return fmt.Errorf("key not found: %s :: %s", n.String(), k.String())
				}
			}
			return nil
		},
	)
	if err == nil {
		return true
	}
	log.Println("Exists:", err)
	return false
}