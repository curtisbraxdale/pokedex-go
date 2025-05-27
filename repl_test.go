package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "MY nAmE iS CuRtIs ",
			expected: []string{"my", "name", "is", "curtis"},
		},
		{
			input:    "Alpha	Beta	Charlie",
			expected: []string{"alpha", "beta", "charlie"},
		},
		{
			input:    "  GETS   RID   OF   CAPS   AND   WHITE   SPACE  ",
			expected: []string{"gets", "rid", "of", "caps", "and", "white", "space"},
		},
		{
			input:    "hopefully this stuff gets me a job prayge",
			expected: []string{"hopefully", "this", "stuff", "gets", "me", "a", "job", "prayge"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		// Check the length of the actual slice against the expected slice
		// if they don't match, use t.Errorf to print an error message
		// and fail the test
		if len(actual) != len(c.expected) {
			t.Errorf("Expected %d words, but got %d", len(c.expected), len(actual))
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			// Check each word in the slice
			// if they don't match, use t.Errorf to print an error message
			// and fail the test
			if word != expectedWord {
				t.Errorf("Expected word %d to be %s, but got %s", i, expectedWord, word)
			}
		}
	}
}
