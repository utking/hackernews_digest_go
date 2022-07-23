package fetcher

import (
	"os"

	"github.com/akamensky/argparse"
)

type ArgParser struct {
	Reverse bool
	Vacuum  bool
	Config  string
}

func (p *ArgParser) Parse() error {
	parser := argparse.NewParser("HackerNews", "Argument parser")

	reverse := parser.Flag("r", "reverse", &argparse.Options{Required: false, Help: "Reverse filters"})
	vacuum := parser.Flag("v", "vacuum", &argparse.Options{Required: false, Help: "Remove old records"})
	config := parser.String("c", "config", &argparse.Options{Required: false,
		Help: "Configuration file", Default: "./config.json"})

	err := parser.Parse(os.Args)
	if err != nil {
		return err
	}

	p.Reverse = *reverse
	p.Vacuum = *vacuum
	p.Config = *config

	return nil
}
