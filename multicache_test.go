package cache

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestNewMulti(t *testing.T) {
	var (
		timesFetched int
	)

	type User struct {
		Username string
		Email    string
	}
	data := []*User{
		{
			Username: "bob",
			Email:    "bob@gmail.com",
		},
		{
			Username: "jen",
			Email:    "jen@yahoo.com",
		},
		{
			Username: "erin",
			Email:    "erin@mail.com",
		},
	}
	mc := NewMulti[string, *User](func(username string) (*User, error) {
		timesFetched++
		for _, user := range data {
			if strings.ToLower(username) == strings.ToLower(user.Username) {
				return user, nil
			}
		}
		return nil, errors.New("not found")
	})
	assert.Equal(t, 0, timesFetched)

	user, err := mc.Get("phil")
	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	assert.Nil(t, user)
	assert.Equal(t, 1, timesFetched)

	user, err = mc.Get("erin")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 2, timesFetched)
	assert.Equal(t, "erin@mail.com", user.Email)

	_, _ = mc.Get("erin")
	_, _ = mc.Get("erin")
	_, _ = mc.Get("erin")
	assert.Equal(t, 2, timesFetched)
}

func TestMultiCache_SetTTLPolicy(t *testing.T) {
	const (
		ttl = 50 * time.Millisecond
	)
	var (
		timesFetched int
	)

	type User struct {
		Username string
		Email    string
	}
	data := []User{
		{
			Username: "bob",
			Email:    "bob@gmail.com",
		},
		{
			Username: "jen",
			Email:    "jen@yahoo.com",
		},
		{
			Username: "erin",
			Email:    "erin@mail.com",
		},
	}
	mc := NewMulti[string, User](func(username string) (User, error) {
		timesFetched++
		for _, user := range data {
			if strings.ToLower(username) == strings.ToLower(user.Username) {
				return user, nil
			}
		}
		return User{}, errors.New("not found")
	})
	assert.Equal(t, 0, timesFetched)
	mc.SetTTLPolicy(ttl)

	_, _ = mc.Get("jen")
	_, _ = mc.Get("jen")
	_, _ = mc.Get("jen")
	assert.Equal(t, 1, timesFetched)
	time.Sleep(2 * ttl)

	_, _ = mc.Get("jen")
	assert.Equal(t, 2, timesFetched, "Jen's user cache should have been invalidated by time to live policy")
}

func TestMultiCache_Invalidate(t *testing.T) {
	var (
		timesFetched     int
		timesInvalidated int
	)

	type User struct {
		Username string
		Email    string
	}
	data := []*User{
		{
			Username: "bob",
			Email:    "bob@gmail.com",
		},
		{
			Username: "jen",
			Email:    "jen@yahoo.com",
		},
		{
			Username: "erin",
			Email:    "erin@mail.com",
		},
	}
	mc := NewMulti[string, *User](func(username string) (*User, error) {
		timesFetched++
		for _, user := range data {
			if strings.ToLower(username) == strings.ToLower(user.Username) {
				return user, nil
			}
		}
		return nil, errors.New("not found")
	})
	mc.OnInvalidate("jen", func() {
		timesInvalidated++
	})
	assert.Equal(t, 0, timesFetched)
	assert.Equal(t, 0, timesInvalidated)

	_, _ = mc.Get("jen")
	_, _ = mc.Get("jen")
	_, _ = mc.Get("jen")
	assert.Equal(t, 1, timesFetched, "only the first search for Jen should increment the fetch count")

	_, _ = mc.Get("bob")
	assert.Equal(t, 2, timesFetched, "searching for Bob should increment the fetch count, since it's a new record")
	mc.Invalidate("bob")

	_, _ = mc.Get("bob")
	assert.Equal(t, 3, timesFetched, "searching for Bob again should increment the fetch count after invalidating this value")

	_, _ = mc.Get("jen")
	assert.Equal(t, 3, timesFetched, "Jen should still be cached")
	assert.Equal(t, 0, timesInvalidated, "Jen should have never been invalidated")
	mc.Invalidate("jen")
	assert.Equal(t, 3, timesFetched, "No load should happen after Invalidate")
	assert.Equal(t, 0, timesInvalidated, "Jen should be considered invalid now")
}

func TestMultiCache_Preheat(t *testing.T) {
	var (
		timesFetched int
		prefetchKeys = []string{"jen", "bob", "erin", "erin"}
	)

	type User struct {
		Username string
		Email    string
	}
	data := []*User{
		{
			Username: "bob",
			Email:    "bob@gmail.com",
		},
		{
			Username: "jen",
			Email:    "jen@yahoo.com",
		},
		{
			Username: "erin",
			Email:    "erin@mail.com",
		},
	}
	mc := NewMulti[string, *User](func(username string) (*User, error) {
		timesFetched++
		for _, user := range data {
			if strings.ToLower(username) == strings.ToLower(user.Username) {
				return user, nil
			}
		}
		return nil, errors.New("not found")
	})
	assert.Equal(t, 0, timesFetched)
	assert.NoError(t, mc.Preheat(prefetchKeys))
	assert.Equal(t, len(prefetchKeys)-1, timesFetched, "A new fetch should have been done for each record, except for duplicates")

	_, _ = mc.Get("jen")
	_, _ = mc.Get("bob")
	_, _ = mc.Get("erin")
	assert.Equal(t, 3, timesFetched, "The records are already fetched and valid, so they should have been returned from cache")
}
