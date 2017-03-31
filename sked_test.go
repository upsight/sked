package sked

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	l := &DiscardLogger{}
	tmpfile, _ := ioutil.TempFile("", "sked")
	defer os.Remove(tmpfile.Name())
	o := &Options{DBPath: tmpfile.Name()}

	s, err := New(o, l)
	defer s.Close()
	if err != nil {
		t.Error(err)
		return
	}
	want := &Options{
		Bucket:           DefaultBucket,
		CheckIntervalMS:  DefaultCheckIntervalMS,
		OpenTimeoutMS:    DefaultOpenTimeoutMS,
		DBPath:           tmpfile.Name(),
		EventChannelSize: DefaultEventChannelSize,
	}
	if !reflect.DeepEqual(want, s.options) {
		t.Errorf("want options %+v, got %+v", want, s.options)
	}
}

func TestAddDeleteGet(t *testing.T) {
	l := NewLogLogger()
	tmpfile, _ := ioutil.TempFile("", "sked")
	defer os.Remove(tmpfile.Name())
	o := &Options{DBPath: tmpfile.Name()}

	s, err := New(o, l)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()
	event := []byte("{some event data}")
	id, err := s.Add(now, event, "tagme", 2)
	if err != nil {
		t.Error(err)
	}
	if id == 0 {
		t.Errorf("want some id > 0, got %d", id)
		return
	}
	r, err := s.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(r.Event, event) {
		t.Errorf("want %s, got %s", event, r.Event)
	}

	err = s.Delete(id, "")
	if err != nil {
		t.Error(err)
	}

	r, err = s.Get(id)
	if err != nil {
		t.Error(err)
	}
	if r != nil {
	}
}

func TestGetAll(t *testing.T) {
	l := &DiscardLogger{}
	tmpfile, _ := ioutil.TempFile("", "sked")
	defer os.Remove(tmpfile.Name())
	o := &Options{DBPath: tmpfile.Name()}

	s, err := New(o, l)
	defer s.Close()
	if err != nil {
		t.Error(err)
		return
	}
	now := time.Now().UTC()
	event := []byte("{some event data}")
	s.Add(now, event, "tagme", 2)
	s.Add(now, event, "tagme", 1)
	s.Add(now, event, "tagme", 3)
	s.Add(now, event, "someothertagme", 3)
	results, err := s.GetAll("tagme")
	if err != nil {
		t.Error(err)
	}
	if len(results) != 3 {
		t.Errorf("want 3 results, got %d", len(results))
	}
}
