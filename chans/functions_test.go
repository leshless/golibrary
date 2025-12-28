package chans_test

import (
	"slices"
	"testing"

	"github.com/leshless/golibrary/chans"
)

func TestReadAll(t *testing.T) {
	testCases := []struct {
		name    string
		getChan func() <-chan int
		result  []int
	}{
		{
			name: "HappyPath",
			getChan: func() <-chan int {
				ch := make(chan int, 10)
				ch <- 1
				ch <- 2
				ch <- 3

				close(ch)

				return ch
			},
			result: []int{1, 2, 3},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ch := testCase.getChan()
			result := chans.ReadAll(ch)

			if slices.Compare(testCase.result, result) != 0 {
				t.Logf("expected: %+v, got: %+v", testCase.result, result)
				t.Fail()
			}
		})
	}
}
