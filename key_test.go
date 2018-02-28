package sked

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestKey_NewKeyString(t *testing.T) {
	now := time.Now().UTC()
	ks := NewKeyString(now, 4, "sometag", 5)
	if !strings.HasSuffix(ks, ";4;sometag;5") {
		t.Errorf("want suffix ';4;sometag;5' got %s", ks)
	}
}

func TestKey_String(t *testing.T) {
	now := time.Now().UTC()
	ks := NewKeyString(now, 4, "sometag", 5)
	k := &Key{Time: now, ID: 4, Tag: "sometag", ExpiresAfter: 5}
	if k.String() != ks {
		t.Errorf("want %s, got %s", ks, k.String())
	}
}

func TestKey_IsExpired(t *testing.T) {
	now := time.Now().UTC()
	key := &Key{}
	tests := []struct {
		name           string
		inTime         time.Time
		inExpiresAfter int
		want           bool
	}{
		{"00", now, 1, false},
		{"01", now.Add(-2 * time.Second), 1, true},
		{"02", now.Add(1 * time.Second), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key.Time = tt.inTime
			key.ExpiresAfter = tt.inExpiresAfter
			if got := key.IsExpired(now); got != tt.want {
				t.Errorf("Key.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_Parse(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name    string
		in      string
		want    *Key
		wantErr error
	}{
		{"00 invalid key", "", &Key{ExpiresAfter: 2, Tag: "abc", Time: now}, ErrInvalidKey},
		{"01", now.Format(time.RFC3339) + ";3;abc;2", &Key{ID: 3, ExpiresAfter: 2, Tag: "abc", Time: now}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKey(tt.in)
			if err != tt.wantErr {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if !reflect.DeepEqual(got.ExpiresAfter, tt.want.ExpiresAfter) {
				t.Errorf("Key.Parse() = %v, want %v", got.ExpiresAfter, tt.want.ExpiresAfter)
			}
			if !reflect.DeepEqual(got.Tag, tt.want.Tag) {
				t.Errorf("Key.Parse() = %v, want %v", got.Tag, tt.want.Tag)
			}
			if !reflect.DeepEqual(got.Time.Format(time.RFC3339), tt.want.Time.Format(time.RFC3339)) {
				t.Errorf("Key.Parse() = %v, want %v", got.Time, tt.want.Time)
			}
		})
	}
}

func TestKey_MatchKey(t *testing.T) {
	tests := []struct {
		name    string
		in      *Key
		want    []byte
		wantErr error
	}{
		{"01 id and tag", &Key{ID: 3, Tag: "abc"}, []byte(";3;abc;"), nil},
		{"01 id only", &Key{ID: 3, Tag: ""}, []byte(";3;"), nil},
		{"02 tag only", &Key{ID: 0, Tag: "def"}, []byte(";def;"), nil},
		{"03 no id or tag", &Key{ID: 0, Tag: ""}, nil, ErrNoMatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.MatchKey()
			if err != tt.wantErr {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("want %s got %s", tt.want, got)
			}
		})
	}
}
