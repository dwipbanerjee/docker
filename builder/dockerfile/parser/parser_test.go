package parser

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const testDir = "testfiles"
const negativeTestDir = "testfiles-negative"

func getDirs(t *testing.T, dir string) []string {
	f, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	dirs, err := f.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}

	return dirs
}

func TestTestNegative(t *testing.T) {
	for _, dir := range getDirs(t, negativeTestDir) {
		dockerfile := filepath.Join(negativeTestDir, dir, "Dockerfile")

		df, err := os.Open(dockerfile)
		if err != nil {
			t.Fatalf("Dockerfile missing for %s: %v", dir, err)
		}

		_, err = Parse(df)
		if err == nil {
			t.Fatalf("No error parsing broken dockerfile for %s", dir)
		}

		df.Close()
	}
}

func TestTestData(t *testing.T) {
	for _, dir := range getDirs(t, testDir) {
		dockerfile := filepath.Join(testDir, dir, "Dockerfile")
		resultfile := filepath.Join(testDir, dir, "result")

		df, err := os.Open(dockerfile)
		if err != nil {
			t.Fatalf("Dockerfile missing for %s: %v", dir, err)
		}
		defer df.Close()

		ast, err := Parse(df)
		if err != nil {
			t.Fatalf("Error parsing %s's dockerfile: %v", dir, err)
		}

		content, err := ioutil.ReadFile(resultfile)
		if err != nil {
			t.Fatalf("Error reading %s's result file: %v", dir, err)
		}

		if runtime.GOOS == "windows" {
			// CRLF --> CR to match Unix behaviour
			content = bytes.Replace(content, []byte{'\x0d', '\x0a'}, []byte{'\x0a'}, -1)
		}

		if ast.Dump()+"\n" != string(content) {
			fmt.Fprintln(os.Stderr, "Result:\n"+ast.Dump())
			fmt.Fprintln(os.Stderr, "Expected:\n"+string(content))
			t.Fatalf("%s: AST dump of dockerfile does not match result", dir)
		}
	}
}

func TestParseWords(t *testing.T) {
	tests := []map[string][]string{
		{
			"input":  {"foo"},
			"expect": {"foo"},
		},
		{
			"input":  {"foo bar"},
			"expect": {"foo", "bar"},
		},
		{
			"input":  {"foo=bar"},
			"expect": {"foo=bar"},
		},
		{
			"input":  {"foo bar 'abc xyz'"},
			"expect": {"foo", "bar", "'abc xyz'"},
		},
		{
			"input":  {`foo bar "abc xyz"`},
			"expect": {"foo", "bar", `"abc xyz"`},
		},
	}

	for _, test := range tests {
		words := parseWords(test["input"][0])
		if len(words) != len(test["expect"]) {
			t.Fatalf("length check failed. input: %v, expect: %v, output: %v", test["input"][0], test["expect"], words)
		}
		for i, word := range words {
			if word != test["expect"][i] {
				t.Fatalf("word check failed for word: %q. input: %v, expect: %v, output: %v", word, test["input"][0], test["expect"], words)
			}
		}
	}
}
