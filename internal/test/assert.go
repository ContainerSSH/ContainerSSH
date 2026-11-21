package test

import (
    "reflect"
    "testing"
)

func AssertEquals[T comparable](t *testing.T, got T, expected T) {
    t.Helper()
    if got != expected {
        t.Fatalf("Unexpected value: %v, expected: %v", got, expected)
    }
}

func AssertNotNil[T comparable](t *testing.T, value T) {
    t.Helper()
    if reflect.ValueOf(value).IsNil() {
        t.Fatalf("Unexpected nil value encountered.")
    }
}

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
}
