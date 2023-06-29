package utils

import (
	"testing"
)

func TestSliceContainsValue(t *testing.T) {
	type testExample[T comparable] struct {
		wantedValue  T
		inputedSlice []T
		output       bool
	}

	tableTest := []testExample[string]{
		{
			wantedValue:  "dog",
			inputedSlice: []string{"foo", "bar", "dog"},
			output:       true,
		},
		{
			wantedValue:  "111",
			inputedSlice: []string{"foo", "bar", "dog"},
			output:       false,
		},
	}

	for _, tt := range tableTest {
		output := SliceContainsValue(tt.wantedValue, tt.inputedSlice)
		if output != tt.output {
			t.Errorf("Searching for %v within %v | got %v, want %v", tt.wantedValue, tt.inputedSlice, output, tt.output)
		}
	}
}

func TestAreSlicesEqual(t *testing.T) {
	type sliceExamples[T comparable] struct {
		sliceOne []T
		sliceTwo []T
		output   bool
	}

	tableTestString := []sliceExamples[string]{
		{
			sliceOne: []string{"dog", "bar", "foo"},
			sliceTwo: []string{"foo", "bar", "dog"},
			output:   true,
		},
		{
			sliceOne: []string{"foo", "bar", "test"},
			sliceTwo: []string{"foo", "bar", "dog"},
			output:   false,
		},
		{
			sliceOne: []string{"bar", "foo", "dog"},
			sliceTwo: []string{"bar", "foo"},
			output:   false,
		},
	}

	for _, tt := range tableTestString {
		output := AreSlicesEqual(tt.sliceOne, tt.sliceTwo)
		if output != tt.output {
			t.Errorf("Comparing slices: %v | %v | got %v, want %v", tt.sliceOne, tt.sliceTwo, output, tt.output)
		}
	}
}

func TestSliceContainsSlice(t *testing.T) {
	type sliceExamples[T comparable] struct {
		containedSlice []T
		toContainSlice []T
		output         bool
	}

	tableTestString := []sliceExamples[string]{
		{
			containedSlice: []string{"bar", "foo"},
			toContainSlice: []string{"foo", "bar", "dog"},
			output:         true,
		},
		{
			containedSlice: []string{"dog", "bar", "foo"},
			toContainSlice: []string{"foo", "bar"},
			output:         false,
		},
		{
			containedSlice: []string{"dog", "bar", "foo"},
			toContainSlice: []string{"foo", "bar", "dog"},
			output:         true,
		},
	}

	for _, tt := range tableTestString {
		output := SliceContainsSlice(tt.containedSlice, tt.toContainSlice)
		if output != tt.output {
			t.Errorf("Does slice %v contains slice %v | got %v, want %v", tt.toContainSlice, tt.containedSlice, output, tt.output)
		}
	}
}
