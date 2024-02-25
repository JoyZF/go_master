package demo

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewEqual(t *testing.T) {
	// Different allocations should not be equal.
	if errors.New("abc") == errors.New("abc") {
		t.Errorf(`New("abc") == New("abc")`)
	}
	if errors.New("abc") == errors.New("xyz") {
		t.Errorf(`New("abc") == New("xyz")`)
	}

	// Same allocation should be equal to itself (not crash).
	err := errors.New("jkl")
	if err != err {
		t.Errorf(`err != err`)
	}
}

func TestJoinReturnsNil(t *testing.T) {
	err2 := errors.New("")
	err2.Error()
	err := errors.Join()
	fmt.Println(err.Error())
}
