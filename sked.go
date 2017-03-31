package sked

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// Defines global options.
const (
	// DefaultBucket is the default bucket name to use if not provided.
	DefaultBucket = "sked"
	// DefaultCheckIntervalMS is the interval to poll the db for new events.
	DefaultCheckIntervalMS = 500
	// DefaultDBPath is the file path to the db.
	DefaultDBPath = "sked.db"
	// DefaultOpenTimeoutMS is the time to wait to open a db.
	DefaultOpenTimeoutMS = 1000
	// DefaultEventChannelSize is the amount of events to buffer on the events channel.
	DefaultEventChannelSize = 1
	// DefaultExpiresAfter is the number of seconds before an event is expired.
	DefaultExpiresAfter = 1800
)

var (
	// ErrNoMatch for cases where there was no match found.
	ErrNoMatch = errors.New("no match found")
)

// Sked ...
type Sked struct {
	DB      *bolt.DB
	done    chan struct{}
	events  chan []byte
	logger  Logger
	options *Options
	rw      *sync.Mutex // boltdb only allows 1 read-write transaction at a time
}

// Options are used to initialize a new Sked.
type Options struct {
	// Bucket is the bucket name to use to find events.
	Bucket string
	// CheckIntervalMS is the interval in milliseconds that the poller
	// should check for scheduled events. Defaults to DefaultCheckIntervalMS.
	CheckIntervalMS int
	// OpenTimeoutMS is the timeout a new Sked will take opening a DB in ms.
	OpenTimeoutMS int
	// DBPath is the path to store the database file.
	DBPath string
	// EventChannelSize is the size of the event channel to buffer before blocking.
	EventChannelSize int
}

// Result is used for GetAll operations.
type Result struct {
	Key   *Key   `json:"key"`
	Event []byte `json:"event"`
}

// New will initialize a new Sked scheduler and start a poller for events.
func New(o *Options, l Logger) (*Sked, error) {
	if o == nil {
		o = &Options{}
	}
	s := &Sked{
		done:    make(chan struct{}),
		logger:  l,
		options: o,
		rw:      &sync.Mutex{},
	}
	if o.Bucket == "" {
		o.Bucket = DefaultBucket
	}
	if o.CheckIntervalMS == 0 {
		o.CheckIntervalMS = DefaultCheckIntervalMS
	}
	if o.OpenTimeoutMS == 0 {
		o.OpenTimeoutMS = DefaultOpenTimeoutMS
	}
	if o.DBPath == "" {
		o.DBPath = DefaultDBPath
	}
	if o.EventChannelSize == 0 {
		o.EventChannelSize = DefaultEventChannelSize
	}
	s.events = make(chan []byte, o.EventChannelSize)
	s.logger.Printf("config: %+v", o)

	db, err := bolt.Open(o.DBPath, 0600, &bolt.Options{Timeout: time.Duration(o.OpenTimeoutMS) * time.Second})
	if err != nil {
		return nil, err
	}
	s.DB = db
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(o.Bucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	go s.poll()
	return s, nil
}

// Add will add a scheduled event. If expiresAfter is not proviced a default will be used.
func (s *Sked) Add(t time.Time, event []byte, tag string, expiresAfter int) (uint64, error) {
	s.rw.Lock()
	defer s.rw.Unlock()
	var id uint64
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.options.Bucket))
		newid, _ := b.NextSequence()
		if expiresAfter == 0 {
			expiresAfter = DefaultExpiresAfter
		}
		keystr := NewKeyString(t, newid, tag, expiresAfter)
		err := b.Put([]byte(keystr), event)
		id = newid
		return err
	})
	return id, err
}

// Delete will remove a scheduled event. You can use id or tag, or both.
func (s *Sked) Delete(id uint64, tag string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.options.Bucket))
		cur := b.Cursor()
		key := &Key{ID: id, Tag: tag}
		matchBytes, err := key.MatchKey()
		if err != nil {
			return err
		}
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			if bytes.Contains(k, matchBytes) {
				b.Delete(k)
			}
		}
		return nil
	})
	return err
}

// Get an event by event id.
func (s *Sked) Get(id uint64) (*Result, error) {
	var (
		result *Result
		err    error
	)
	err = s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.options.Bucket))
		cur := b.Cursor()
		key := &Key{ID: id}
		matchBytes, err := key.MatchKey()
		if err != nil {
			return err
		}
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if bytes.Contains(k, matchBytes) {
				r := &Result{Event: v}
				r.Key, _ = key.Parse(string(k))
				result = r
				return nil
			}
		}
		return nil
	})
	return result, err
}

// GetAll will get all events by tag.
func (s *Sked) GetAll(tag string) ([]*Result, error) {
	var (
		results = []*Result{}
		err     error
	)
	err = s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.options.Bucket))
		cur := b.Cursor()
		key := &Key{Tag: tag}
		matchBytes, err := key.MatchKey()
		if err != nil {
			return err
		}
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if bytes.Contains(k, matchBytes) {
				r := &Result{Event: v}
				r.Key, _ = key.Parse(string(k))
				results = append(results, r)
			}
		}
		return nil
	})
	return results, err
}

// Close will close the DB.
func (s *Sked) Close() error {
	s.done <- struct{}{}
	return s.DB.Close()
}
