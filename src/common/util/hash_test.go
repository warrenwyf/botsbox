package util

import (
	"testing"
)

type object struct {
	name   string
	Params map[string]interface{}
}

func (o *object) GetName() string {
	return o.name
}

func Test_Md5_string(t *testing.T) {
	a := "Hello World"
	b := "Hello World!"

	if Md5(a) == Md5(b) {
		t.Fatalf("Md5() failed: %v == %v", a, b)
	}
}

func Test_Md5_map(t *testing.T) {
	a := map[string]interface{}{
		"name": 1,
	}
	b := map[string]interface{}{
		"name": 2,
	}

	if Md5(a) == Md5(b) {
		t.Fatalf("Md5() failed: %v == %v", a, b)
	}
}

func Test_Md5_Equal(t *testing.T) {
	a := object{
		name: "a",
		Params: map[string]interface{}{
			"foo": 0,
			"bar": 1,
		},
	}

	b := object{
		Params: map[string]interface{}{
			"bar": 1,
			"foo": 0,
		},
		name: "b",
	}

	if Md5(a) != Md5(b) {
		t.Fatalf("Md5() failed: %v == %v", a, b)
	}
}

func Test_Md5_NotEqual(t *testing.T) {
	a := object{
		name: "a",
		Params: map[string]interface{}{
			"foo": 0,
			"bar": 1,
		},
	}

	b := object{
		name: "b",
		Params: map[string]interface{}{
			"foo": "0",
			"bar": 1,
		},
	}

	if Md5(a) == Md5(b) {
		t.Fatalf("Md5() failed: %v != %v", a, b)
	}
}
