package fetcher

import (
	"os"
	"testing"
)

func TestArgParseEmptySet(t *testing.T) {
	prevArgs := os.Args
	newArgs := []string{"self"}
	os.Args = newArgs
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
	os.Args = prevArgs
}

func TestArgParseValueSet(t *testing.T) {
	prevArgs := os.Args
	newArgs := []string{"self", "-r", "-v", "-c", "another-config.json"}

	os.Args = newArgs

	args := ArgParser{}

	if err := args.Parse(); err != nil {
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
	os.Args = prevArgs
}
