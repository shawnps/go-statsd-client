package statsd

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

var statsdSubStatterPacketTests = []struct {
	Prefix    string
	SubPrefix string
	Method    string
	Stat      string
	Value     interface{}
	Rate      float32
	Expected  string
}{
	{"test", "sub", "Gauge", "gauge", int64(1), 1.0, "test.sub.gauge:1|g"},
	{"test", "sub", "Inc", "count", int64(1), 0.999999, "test.sub.count:1|c|@0.999999"},
	{"test", "sub", "Inc", "count", int64(1), 1.0, "test.sub.count:1|c"},
	{"test", "sub", "Dec", "count", int64(1), 1.0, "test.sub.count:-1|c"},
	{"test", "sub", "Timing", "timing", int64(1), 1.0, "test.sub.timing:1|ms"},
	{"test", "sub", "TimingDuration", "timing", 1500 * time.Microsecond, 1.0, "test.sub.timing:1.5|ms"},
	{"test", "sub", "TimingDuration", "timing", 3 * time.Microsecond, 1.0, "test.sub.timing:0.003|ms"},
	{"test", "sub", "Set", "strset", "pickle", 1.0, "test.sub.strset:pickle|s"},
	{"test", "sub", "SetInt", "intset", int64(1), 1.0, "test.sub.intset:1|s"},
	{"test", "sub", "GaugeDelta", "gauge", int64(1), 1.0, "test.sub.gauge:+1|g"},
	{"test", "sub", "GaugeDelta", "gauge", int64(-1), 1.0, "test.sub.gauge:-1|g"},
}

func TestSubStatterClient(t *testing.T) {
	l, err := newUDPListener("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for _, tt := range statsdSubStatterPacketTests {
		c, err := NewClient(l.LocalAddr().String(), tt.Prefix)
		if err != nil {
			t.Fatal(err)
		}
		s := c.NewSubStatter(tt.SubPrefix)
		method := reflect.ValueOf(s).MethodByName(tt.Method)
		e := method.Call([]reflect.Value{
			reflect.ValueOf(tt.Stat),
			reflect.ValueOf(tt.Value),
			reflect.ValueOf(tt.Rate)})[0]
		errInter := e.Interface()
		if errInter != nil {
			t.Fatal(errInter.(error))
		}

		data := make([]byte, 128)
		_, _, err = l.ReadFrom(data)
		if err != nil {
			c.Close()
			t.Fatal(err)
		}

		data = bytes.TrimRight(data, "\x00")
		if bytes.Equal(data, []byte(tt.Expected)) != true {
			c.Close()
			t.Fatalf("%s got '%s' expected '%s'", tt.Method, data, tt.Expected)
		}
		c.Close()
	}
}

func TestSubStatterClosedClient(t *testing.T) {
	l, err := newUDPListener("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for _, tt := range statsdSubStatterPacketTests {
		c, err := NewClient(l.LocalAddr().String(), tt.Prefix)
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
		s := c.NewSubStatter(tt.SubPrefix)
		method := reflect.ValueOf(s).MethodByName(tt.Method)
		e := method.Call([]reflect.Value{
			reflect.ValueOf(tt.Stat),
			reflect.ValueOf(tt.Value),
			reflect.ValueOf(tt.Rate)})[0]
		errInter := e.Interface()
		if errInter == nil {
			t.Fatal("Expected error but got none")
		}
	}
}

func TestNilSubStatterClient(t *testing.T) {
	l, err := newUDPListener("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	/*
		var c *Client
		_, err = c.NewSubStatter("failboat")
		if err == nil {
			t.Fatal("Unexpected lack of error")
		}
	*/
	for _, tt := range statsdSubStatterPacketTests {
		var c *Client
		s := c.NewSubStatter(tt.SubPrefix)

		method := reflect.ValueOf(s).MethodByName(tt.Method)
		e := method.Call([]reflect.Value{
			reflect.ValueOf(tt.Stat),
			reflect.ValueOf(tt.Value),
			reflect.ValueOf(tt.Rate)})[0]
		errInter := e.Interface()
		if errInter != nil {
			t.Fatal(errInter.(error))
		}

		data := make([]byte, 128)
		n, _, err := l.ReadFrom(data)
		// this is expected to error, since there should
		// be no udp data sent, so the read will time out
		if err == nil || n != 0 {
			c.Close()
			t.Fatal(err)
		}
		c.Close()
	}
}

func TestNoopSubStatterClient(t *testing.T) {
	l, err := newUDPListener("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for _, tt := range statsdSubStatterPacketTests {
		c, err := NewNoopClient(l.LocalAddr().String(), tt.Prefix)
		if err != nil {
			t.Fatal(err)
		}
		s := c.NewSubStatter(tt.SubPrefix)
		method := reflect.ValueOf(s).MethodByName(tt.Method)
		e := method.Call([]reflect.Value{
			reflect.ValueOf(tt.Stat),
			reflect.ValueOf(tt.Value),
			reflect.ValueOf(tt.Rate)})[0]
		errInter := e.Interface()
		if errInter != nil {
			t.Fatal(errInter.(error))
		}

		data := make([]byte, 128)
		n, _, err := l.ReadFrom(data)
		// this is expected to error, since there should
		// be no udp data sent, so the read will time out
		if err == nil || n != 0 {
			c.Close()
			t.Fatal(err)
		}
		c.Close()
	}
}
