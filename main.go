package main

import (
	"fmt"
	"os"

	"github.com/angrycub/parse-nomad-config/version"
	"github.com/hashicorp/nomad/command"
	"github.com/hashicorp/nomad/command/agent"
	flag "github.com/spf13/pflag"
)

var (
	jsonFlag       = flag.BoolP("json", "j", false, "output in JSON format")
	templateString = flag.StringP("template", "t", "", "Go template to format the output with")
	outputFile     = flag.StringP("out", "o", "", "write to the given file, instead of stdout")
	showVersion    = flag.BoolP("version", "v", false, "show the version number and immediately exit")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion().FullVersionNumber(true))
		os.Exit(0)
	}

	args := flag.Args()

	err := realmain(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		os.Exit(1)
	}

}

func realmain(args []string) error {
	var config *agent.Config
	var err error
	if len(args) == 0 {
		return fmt.Errorf("no configuration file supplied")
	}
	for _, a := range args {
		config, err = agent.LoadConfig(a)
		if err != nil {
			return fmt.Errorf("failed to load configuration. file:%s error:%s", a, err)
		}
	}
	var out string
	if *jsonFlag || len(*templateString) > 0 {
		out, err = command.Format(*jsonFlag, *templateString, config)
		if err != nil {
			return fmt.Errorf("failed outputting config. err:%s", err)
		}
	}

	target := os.Stdout
	if *outputFile != "" {
		target, err = os.OpenFile(*outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("can't open %s for writing: %s", *outputFile, err)
		}
	}

	fmt.Fprintf(target, "%s\n", out)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: parse-nomad-config [options] [hcl-file ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
