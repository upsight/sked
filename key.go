package sked

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// Sentinel is used to create keys.
	Sentinel = ";"
)

var (
	// ErrInvalidKey is for invalid key strings.
	ErrInvalidKey = errors.New("invalid key")
)

// Key represents the keys stored in the db.
type Key struct {
	Time         time.Time `json:"time"`
	ID           uint64    `json:"id"`
	ExpiresAfter int       `json:"expires_after"`
	Tag          string    `json:"tag"`
}

// NewKeyString will create a new key for Add. The format is date;id;tag;expires_after
func NewKeyString(t time.Time, id uint64, tag string, expiresAfter int) string {
	args := []string{t.Format(time.RFC3339), strconv.FormatUint(id, 10), tag, strconv.Itoa(expiresAfter)}
	return strings.Join(args, Sentinel)
}

// ParseKey will parse the Key from an input string.
func ParseKey(in string) (*Key, error) {
	tokens := strings.Split(in, Sentinel)
	if len(tokens) < 4 {
		return nil, ErrInvalidKey
	}
	key := &Key{
		Tag: tokens[2],
	}
	key.ID, _ = strconv.ParseUint(tokens[1], 10, 64)
	key.Time, _ = time.Parse(time.RFC3339, tokens[0])
	key.ExpiresAfter, _ = strconv.Atoi(tokens[3])
	return key, nil
}

// String returns the string representation of a Key.
func (k *Key) String() string {
	args := []string{k.Time.Format(time.RFC3339), strconv.FormatUint(k.ID, 10), k.Tag, strconv.Itoa(k.ExpiresAfter)}
	return strings.Join(args, Sentinel)
}

// IsExpired checks if this key is older than the ExpiresAfter field.
func (k *Key) IsExpired(now time.Time) bool {
	when := k.Time.Add(time.Duration(k.ExpiresAfter) * time.Second)
	if when.Before(now) {
		return true
	}
	return false
}

// Match will create a match key for a given id, tag combination.
func (k *Key) MatchKey() ([]byte, error) {
	match := ""
	switch {
	case k.ID != 0 && k.Tag == "":
		// ;id;
		match = fmt.Sprintf("%s%d%s", Sentinel, k.ID, Sentinel)
	case k.ID == 0 && k.Tag != "":
		// ;tag;
		match = fmt.Sprintf("%s%s%s", Sentinel, k.Tag, Sentinel)
	case k.ID != 0 && k.Tag != "":
		// ;id;tag;
		match = fmt.Sprintf("%s%d%s%s%s", Sentinel, k.ID, Sentinel, k.Tag, Sentinel)
	default:
		// no-op
	}
	if match == "" {
		return nil, ErrNoMatch
	}
	return []byte(match), nil
}
