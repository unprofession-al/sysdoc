package main

import (
	"reflect"
	"testing"
)

func TestGetPosition(t *testing.T) {
	tests := []struct {
		basepath, path string
		expected       []string
	}{
		{"/root", "/root/file.txt", []string{"file.txt"}},
		{"/root", "/root/dir1/dir2/file.txt", []string{"dir1", "dir2", "file.txt"}},
		{"/root", "/file.txt", []string{"file.txt"}},
	}

	for _, tc := range tests {
		result := getPosition(tc.basepath, tc.path)
		if len(result) != len(tc.expected) {
			t.Errorf("Expected %v, but got %v", tc.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
				break
			}
		}
	}
}

func TestPositionFromID(t *testing.T) {
	tests := []struct {
		input string
		sep   string
		want  []string
	}{
		{input: "a.b.c", sep: ".", want: []string{"a", "b", "c"}},
		{input: "a.b.c.", sep: ".", want: []string{"a", "b", "c"}},
		{input: ".a.b.c", sep: ".", want: []string{"a", "b", "c"}},
		{input: "abc", sep: ".", want: []string{"abc"}},
		{input: "a/b/c", sep: "/", want: []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		got := positionFromID(tc.input, tc.sep)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
