package main

import (
	"fmt"
	"log"
	"time"

	"github.com/upsight/sked"
)

func main() {
	s, err := sked.New(nil, sked.NewLogLogger())
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Listen for events
	go func() {
		for e := range s.Events() {
			log.Println("Got event", string(e))
		}
	}()

	// Add an event
	id, err := s.Add(time.Now().UTC().Add(time.Duration(1)*time.Second), []byte("test"), "mytag", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Added", id)
	// Get added event by id
	got, err := s.Get(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get %d %+v %s\n", id, got.Key, got.Event)

	// Get added events by tag
	gottag, err := s.GetAll("mytag")
	if err != nil {
		log.Fatal(err)
	}
	for _, g := range gottag {
		fmt.Println("GetAll mytag", g.Key, string(g.Event))
	}

	// Add another event
	otherid, err := s.Add(time.Now().UTC().Add(time.Duration(1)*time.Second), []byte("testother"), "myothertag", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Added", otherid)

	// Get by id
	got, err = s.Get(otherid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get %d %+v %s\n", otherid, got.Key, got.Event)

	// Get by tag
	gottag, err = s.GetAll("myothertag")
	if err != nil {
		log.Fatal(err)
	}
	for _, g := range gottag {
		fmt.Println("GetAll myothertag", g.Key, string(g.Event))
	}

	// Delete myothertag
	err = s.Delete(0, "myothertag")
	if err != nil {
		log.Fatal(err)
	}
	gottag, err = s.GetAll("myothertag")
	if err != nil {
		log.Fatal(err)
	}
	for _, g := range gottag {
		fmt.Println("GetAll myothertag", g.Key, string(g.Event))
	}
	time.Sleep(time.Second)
}
