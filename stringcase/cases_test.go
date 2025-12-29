package stringcase_test

import (
	"testing"

	"github.com/leshless/golibrary/stringcase"
)

func TestLowerCamel(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"hello_my_name_is_artem",
			"helloMyNameIsArtem",
		},
		{
			"BTRFS Is a modern file system",
			"btrfsIsAModernFileSystem",
		},
		{
			"Zooweemama",
			"zooweemama",
		},
		{
			"EntityID",
			"entityID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			result := stringcase.LowerCamel(testCase.input)
			if result != testCase.output {
				t.Logf("expected: %s, got: %s", testCase.output, result)
				t.Fail()
			}
		})
	}
}

func TestUpperCamel(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"hello_my_name_is_artem",
			"HelloMyNameIsArtem",
		},
		{
			"BTRFS Is a modern file system",
			"BTRFSIsAModernFileSystem",
		},
		{
			"Zooweemama",
			"Zooweemama",
		},
		{
			"Powerman6000",
			"Powerman6000",
		},
		{
			"entity ID",
			"EntityID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			result := stringcase.UpperCamel(testCase.input)
			if result != testCase.output {
				t.Logf("expected: %s, got: %s", testCase.output, result)
				t.Fail()
			}
		})
	}
}

func TestLowerSnake(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"hello_my_name_is_artem",
			"hello_my_name_is_artem",
		},
		{
			"BTRFS Is a modern file system",
			"btrfs_is_a_modern_file_system",
		},
		{
			"Zooweemama",
			"zooweemama",
		},
		{
			"Powerman6000",
			"powerman6000",
		},
		{
			"entity ID",
			"entity_id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			result := stringcase.LowerSnake(testCase.input)
			if result != testCase.output {
				t.Logf("expected: %s, got: %s", testCase.output, result)
				t.Fail()
			}
		})
	}
}
