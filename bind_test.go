package hr

import (
	"testing"
)

type structAll struct {
	Str      string  `query:"str"`
	I64      int64   `query:"i64"`
	SliceI64 []int64 `query:"slice_i64"`
	Time     Time    `query:"time"`
	IP       IP      `query:"ip"`

	PtrStr  *string `query:"ptr_str"`
	PtrI64  *int64  `query:"ptr_i64"`
	PtrTime *Time   `query:"ptr_time"`
	PtrIP   *IP     `query:"ptr_ip"`

	SliceTime []Time `query:"slice_time"`
	SliceIP   []IP   `query:"slice_ip"`
}

var valuesAll = map[string][]string{
	"str":       {"what is the ultimate answer to the universe?"},
	"i64":       {"42"},
	"slice_i64": {"142", "242", "342"},
	"time":      {"2023-06-05T21:33:45+08:00"},
	"ip":        {"127.0.0.1"},

	"ptr_str":  {"what is the ultimate answer to the universe?"},
	"ptr_i64":  {"42"},
	"ptr_time": {"2023-06-05T21:33:45+08:00"},
	"ptr_ip":   {"127.0.0.1"},

	"slice_time": {"2023-06-05T21:33:45+08:00", "2023-06-06T15:29:56+08:00"},
	"slice_ip":   {"127.0.0.1", "2001:db8::68"},
}

func TestBind(t *testing.T) {
	var a structAll
	if err := bind(&a, valuesAll, "query"); err != nil {
		t.Fatal(err)
	}
}

type structOptional struct {
	Name string `form:"name"`
	Age  string `form:"age?"` // optional
}

func TestBindOptional(t *testing.T) {
	var a structOptional

	v1 := map[string][]string{}
	err := bind(&a, v1, "form")
	if err == nil || err.Error() != "missing field: name" {
		t.Fatal("unexpected error")
	}

	v2 := map[string][]string{"name": {"Bob"}}
	err = bind(&a, v2, "form")
	if err != nil {
		t.Fatal(err)
	}
}
