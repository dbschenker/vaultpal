package cache

import (
	"encoding/gob"
	"os"
	"time"
)

type Cache struct {
	Address string
	Token   string
	Updated time.Time
	TTL     time.Duration
}

var entries = make(map[string]Cache)
var cacheFile = os.TempDir() + "/vaultpal-timer.gob"

func Clear() {
	_ = os.Remove(cacheFile)
}

func get() (e map[string]Cache, err error) {
	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)

	err = decoder.Decode(&entries)
	if err != nil {
		return nil, err
	}

	if err := file.Close(); err != nil {
		return nil, err
	}
	return entries, nil
}

func Write(key string, c Cache) error {
	e, err := get()
	if err != nil {
		e = entries
	}

	file, err := os.Create(cacheFile)
	if err != nil {
		return err
	}

	e[key] = c
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(e)
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}
	return err
}

func Read(key string) (Cache, error) {
	e, err := get()
	if err != nil {
		return Cache{}, err
	}

	return e[key], nil

}
