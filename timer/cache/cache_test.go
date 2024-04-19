package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	to := "1234"
	ep := "localhost:1234"
	c := Cache{
		Address: ep,
		Token:   to,
		Updated: time.Now(),
		TTL:     1 * time.Second,
	}
	assert.NoError(t, Write(ep, c))
	defer Clear()
	c2, err := Read(ep)
	assert.NoError(t, err)
	assert.Equal(t, to, c2.Token)
	assert.Equal(t, ep, c2.Address)

}
