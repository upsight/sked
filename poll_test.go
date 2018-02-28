package sked

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

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
