package helpers

import (
	h "k.prv/rpimon/helpers"
	"testing"
)

func TestRingBuff1(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	if rbuff.Len() != 0 {
		t.Errorf("Invalid size %d (should be 0)", rbuff.Len())
	}
	idx := rbuff.Put(0)
	if rbuff.Len() != 1 {
		t.Errorf("Invalid size %d (should be 1)", idx)
	}

	if val, ok := rbuff.Get(0); !ok || val != 0 {
		t.Errorf("Invalid value %s (should be 0) - %s", val, rbuff.String())
	}

	for i := 1; i < 5; i++ {
		rbuff.Put(i)
		t.Log(rbuff.String())
	}

	if rbuff.Len() != 5 {
		t.Errorf("Invalid size %s (should be 5)", rbuff.String())
	}

	for i := 0; i < 5; i++ {
		if val, ok := rbuff.Get(i); !ok || val != i {
			t.Errorf("Invalid value on idx=%d - %v, %s", i, val, rbuff.String())
		}
	}
}

func TestRingBuff2(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	if rbuff.Len() != 0 {
		t.Errorf("Invalid size %d (should be 0)", rbuff.Len())
	}

	for i := 0; i < 20; i++ {
		rbuff.Put(i)
		t.Log(rbuff.String())
	}

	if rbuff.Len() != 10 {
		t.Errorf("Invalid size %s (should be 1)", rbuff.String())
	}

	for i := 0; i < 10; i++ {
		if val, ok := rbuff.Get(i); !ok || val != i+10 {
			t.Errorf("Invalid value on idx=%d - %v, %s", i, val, rbuff.String())
		}
	}
}

func TestRingBuff3(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	if rbuff.Len() != 0 {
		t.Errorf("Invalid size %d (should be 0)", rbuff.Len())
	}

	for i := 0; i < 55; i++ {
		rbuff.Put(i)
	}

	if rbuff.Len() != 10 {
		t.Errorf("Invalid size %d (should be 10)", rbuff.Len())
	}

	for i := 0; i < 10; i++ {
		expected := i + 45
		if val, ok := rbuff.Get(i); !ok || val != expected {
			t.Errorf("Invalid value on idx=%d - %v (e %d), %s", i, val, expected, rbuff.String())
		}
	}
}

func TestRingBuffSlice(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	for i := 0; i < 55; i++ {
		rbuff.Put(i)
	}

	slice := rbuff.ToSlice()
	if slice == nil || len(slice) != 10 {
		t.Errorf("Invalid slice:%v, %s", slice, rbuff.String())
	}

	for i := 0; i < 10; i++ {
		expected := i + 45
		if slice[i] != expected {
			t.Errorf("Invalid value on idx=%d - %v (e %d)", i, slice[i], expected)
		}
	}
}

func TestRingBuffStrSlice(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	for i := 0; i < 55; i++ {
		rbuff.Put(string(i))
	}

	slice := rbuff.ToStringSlice()

	for i := 0; i < 10; i++ {
		expected := string(i + 45)
		if slice[i] != expected {
			t.Errorf("Invalid value on idx=%d - %v (e %s)", i, slice[i], expected)
		}
	}
}

func TestRingBuffStrSlice2(t *testing.T) {
	rbuff := h.NewRingBuffer(10)
	for i := 0; i < 9; i++ {
		rbuff.Put(string(i + 32))
	}
	t.Logf("Buff: %#v", rbuff.String())

	slice := rbuff.ToStringSlice()
	t.Logf("Slice: %#v", slice)

	for i := 0; i < 9; i++ {
		expected := string(i + 32)
		if slice[i] != expected {
			t.Errorf("Invalid value on idx=%d - %v (e %s)", i, slice[i], expected)
		}
	}
}
