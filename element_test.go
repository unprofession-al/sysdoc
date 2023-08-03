package main

import (
    "testing"
    "reflect"
)

func TestPositionFromID(t *testing.T) {
	  tests := []struct {
        input string
        sep   string
        want  []string
    }{
        {input: "a.b.c", sep: ".", want: []string{"a", "b", "c"}},
        {input: "a.b.c.", sep: ".", want: []string{"a","b","c"}},
        {input: ".a.b.c", sep: ".", want: []string{"a","b","c"}},
        {input: "abc", sep: ".", want: []string{"abc"}},
        {input: "a/b/c", sep: "/", want: []string{"a", "b", "c" }},
    }
        for _, tc := range tests {
        got := positionFromID(tc.input, tc.sep)
        if !reflect.DeepEqual(tc.want, got) {
            t.Fatalf("expected: %v, got: %v", tc.want, got)
        }
    }
}
