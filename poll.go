package sked

import (
	"time"

	"github.com/boltdb/bolt"
)

var (
	// MinDate is the minimum date to search for.
	MinDate = []byte("1900-01-01T00:00:00Z")
)

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
				max := now.Add(time.Second)
			LOOP:
				for k, v := cur.First(); k != nil; k, v = cur.Next() {
					key, err := ParseKey(string(k))
					if err != nil {
						s.logger.Errorln(err)
						continue LOOP
					}
					if key.IsExpired(now) {
						if err := bucket.Delete(k); err != nil {
							s.logger.Errorln(err)
						}
						continue LOOP
					}

					// test if this key is in the future, since keys are sorted we can
					// just break here.
					if key.Time.After(max) {
						break LOOP
					}
					s.events <- v
					if err := bucket.Delete(k); err != nil {
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
