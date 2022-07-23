package fetcher

import (
	"os"
	"testing"
)

func TestArgParseEmptySet(t *testing.T) {
	prev_args := os.Args
	new_args := make([]string, 0)
	new_args = append(new_args, "self")
	os.Args = new_args
	args := ArgParser{}
	err := args.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if args.Config != "./config.json" {
		t.Fatal("--config was not set, and should be the default")
	}
	if args.Reverse {
		t.Fatal("--reverse was not set, should be false")
	}
	if args.Vacuum {
		t.Fatal("--vacuum was not set, should be false")
	}
	// Restore the old Args
	os.Args = prev_args
}

func TestArgParseValueSet(t *testing.T) {
	prev_args := os.Args
	new_args := make([]string, 0)
	new_args = append(new_args, "self")
	new_args = append(new_args, "-r")
	new_args = append(new_args, "-v")
	new_args = append(new_args, "-c")
	new_args = append(new_args, "another-config.json")
	os.Args = new_args
	args := ArgParser{}
	err := args.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if args.Config != "another-config.json" {
		t.Fatal("--config was not set, and should be the default")
	}
	if !args.Reverse {
		t.Fatal("--reverse was set, should be true")
	}
	if !args.Vacuum {
		t.Fatal("--vacuum was set, should be true")
	}
	// Restore the old Args
	os.Args = prev_args
}
