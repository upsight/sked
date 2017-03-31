package sked

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSked_isExpired(t *testing.T) {
	now := time.Now().UTC()
	s := &Sked{}
	tests := []struct {
		name        string
		inTime      time.Time
		inKeyString string
		want        bool
	}{
		{"00", now, "", true},
		{"01", now, NewKeyString(now.Add(-1*time.Second), 1, "", 1), true},
		{"02", now, NewKeyString(now.Add(1*time.Second), 1, "", 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.inTime)
			if got := s.isExpired(tt.inTime, tt.inKeyString); got != tt.want {
				t.Errorf("Sked.isExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSked_poll(t *testing.T) {
	l := NewLogLogger()
	tmpfile, _ := ioutil.TempFile("", "sked")
	defer os.Remove(tmpfile.Name())
	o := &Options{DBPath: tmpfile.Name(), CheckIntervalMS: 10}

	s, err := New(o, l)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().UTC()
	event := []byte("{some event data}")
	s.Add(now, event, "tagme", 2)

	go func() {
		ev := <-s.Events()
		if !bytes.Equal(ev, event) {
			t.Fatalf("want %s, got %s", event, ev)
		}
	}()
	time.Sleep(15 * time.Millisecond)
}
