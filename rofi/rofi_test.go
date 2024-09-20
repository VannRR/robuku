package rofi

import (
	"encoding/ascii85"
	"fmt"
	"os"
	"strings"
	"testing"
)

type MockData struct {
	key   string
	value string
}

func (d *MockData) Bytes() []byte {
	return []byte(fmt.Sprintf("%s%s%s",
		d.key,
		"\u2028",
		d.value,
	))
}

func (d *MockData) ParseBytes(b []byte) error {
	s := string(b)
	vals := strings.Split(s, "\u2028")
	d.key = vals[0]
	d.value = vals[1]
	return nil
}

func Test_getData(t *testing.T) {
	r, err := NewRofiApi[*MockData](&MockData{})
	if err != nil {
		t.Fatalf("expected no error from NewRofiApi(), got %v", err)
	}
	r.Data = &MockData{"1", "1"}

	expected := &MockData{"foo", "bar"}
	bytes := expected.Bytes()
	encodedValue := make([]byte, ascii85.MaxEncodedLen(len(bytes)))
	ascii85.Encode(encodedValue, bytes)

	os.Setenv(dataEnvVar, string(encodedValue))
	err = r.getData()
	if err != nil {
		t.Fatalf("expected no error from getData(), got %v", err)
	}

	if r.Data.key != "foo" {
		t.Errorf("expected key 'foo', got '%v'", r.Data.key)
	}
	if r.Data.value != "bar" {
		t.Errorf("expected value 'bar', got '%v'", r.Data.value)
	}
}

func Test_setData(t *testing.T) {
	os.Setenv(dataEnvVar, "")
	r, err := NewRofiApi[*MockData](&MockData{})
	if err != nil {
		t.Fatalf("expected no error from NewRofiApi(), got %v", err)
	}
	r.Data = &MockData{"1", "1"}
	err = r.setData()
	if err != nil {
		t.Fatalf("expected no error from setData(), got %v", err)
	}

	bytes := r.Data.Bytes()
	expected := make([]byte, ascii85.MaxEncodedLen(len(bytes)))
	ascii85.Encode(expected, bytes)

	actual := r.Options["data"]

	if string(expected) != actual {
		t.Errorf("expected dataEnv value '%s', got '%s'", string(expected), actual)
	}
}
