package sked

import (
	"bytes"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	// MinDate is the minimum date to search for.
	MinDate = []byte("1900-01-01T00:00:00Z")
)

// isExpired parses the key and check if the expires after is valid.
func (s *Sked) isExpired(now time.Time, k string) bool {
	key := &Key{}
	keyP, err := key.Parse(k)
	if err != nil {
		return true
	}
	if keyP.ExpiresAfter > 0 {
		when := keyP.Time.Add(time.Duration(keyP.ExpiresAfter) * time.Second)
		if when.Before(now) {
			return true
		}
	}
	return false
}

// poll will periodically check for events at the specified Options interval.
func (s *Sked) poll() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(s.options.CheckIntervalMS))
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.rw.Lock()
			err := s.DB.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(s.options.Bucket))
				cur := bucket.Cursor()

				now := time.Now().UTC()
				max := []byte(now.Format(time.RFC3339))
				for k, v := cur.Seek(MinDate); k != nil && bytes.HasPrefix(k[:len(time.RFC3339)], max) == true; k, v = cur.Next() {
					// check for expired events
					if s.isExpired(now, string(k)) {
						s.logger.Errorln(fmt.Errorf("found expired key %s %s", k, v))
					} else {
						s.events <- v
					}

					err := bucket.Delete(k)
					if err != nil {
						s.logger.Errorln(err)
					}
				}
				return nil
			})
			s.rw.Unlock()
			if err != nil {
				s.logger.Errorln(err)
			}
		}
	}
}

// Events returns a channel of events as they occur.
func (s *Sked) Events() chan []byte {
	return s.events
}
