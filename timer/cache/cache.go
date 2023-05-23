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

var cacheFile = os.TempDir() + "/token-timer.gob"

func Clear() {
	_ = os.Remove(cacheFile)
}

func WriteCache(object interface{}) error {
	file, err := os.Create(cacheFile)
	if err == nil {
		encoder := gob.NewEncoder(file)
		_ = encoder.Encode(object)
	}
	file.Close()
	return err
}

func ReadCache(object interface{}) error {
	file, err := os.Open(cacheFile)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
