package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jbowes/oag/config"
	"github.com/jbowes/oag/mutator"
	"github.com/jbowes/oag/openapi"
	"github.com/jbowes/oag/translator"
	"github.com/jbowes/oag/writer"
)

var cfgFile = flag.String("c", ".oag.yaml", "Use this configuration file.")

func usage() {
	fmt.Println("Usage: oag [init]")
	os.Exit(-1)
}

func initConfig() error {
	f, err := os.OpenFile(*cfgFile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	defer f.Close()
	if err = config.WriteDefaultConfig(f); err != nil {
		return err
	}

	fmt.Printf("Wrote '%s'. Please edit it for your configuration.\n", *cfgFile)
	return nil
}

func generateClient() error {
	cfg, err := config.Load(*cfgFile)
	if err != nil {
		return err
	}

	doc, err := openapi.LoadFile(cfg.Document)
	if err != nil {
		return err
	}

	code, err := translator.Translate(doc, cfg.Package.Path, cfg.Package.Name, cfg.Types, cfg.StringFormats)
	if err != nil {
		return err
	}

	code = mutator.Mutate(code)

	o, _ := ioutil.ReadFile(cfg.Output)

	var buf bytes.Buffer
	if err = writer.Write(&buf, code, &cfg.Boilerplate); err != nil {
		return err
	}

	n := buf.Bytes()
	if bytes.Equal(n, o) {
		fmt.Println("No change to", cfg.Output)
		return nil
	}

	if err = ioutil.WriteFile(cfg.Output, n, 0666); err != nil {
		return err
	}

	fmt.Println("Wrote", cfg.Output)
	return nil
}

func main() {

	flag.Parse()
	args := flag.Args()

	switch {
	case len(args) == 0:
		if err := generateClient(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	case len(args) > 1, args[0] != "init":
		usage()
	default:
		if err := initConfig(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

}
