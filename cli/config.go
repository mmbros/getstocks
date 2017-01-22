// Package cli provides a CLI UI for the gentmpl command line tool.
package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

const (
	// name of the command line parameters
	clConfig    = "c"
	clDebug     = "d"
	clGenConfig = "g"
	clHelp      = "h"
	clOutput    = "o"

	// default values
	defaultConfigFile = "getstocks.cfg"
	defaultOutputFile = "" // if empty use StdOut
)

type clArgs struct {
	config    string
	debug     bool
	genConfig bool
	help      bool
	output    string
}

type configScraper struct {
	Disabled bool
	Workers  int
}

type configStockSource struct {
	ScraperKey string
	URL        string
}

type configStock struct {
	Isin        string
	Name        string
	Description string
	Sources     []configStockSource `toml:"sources"`
}

type config struct {
	Scrapers map[string]configScraper `toml:"scrapers"`
	Stocks   map[string]configStock   `toml:"stocks"`
}

func unmarshalConfig(data []byte) (*config, error) {
	// parse config file
	var cfg config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// loadConfigFromFile returns the configuration from a configuration file
func loadConfigFromFile(path string) (*config, error) {
	// open config file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// read config file
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return unmarshalConfig(buf)
}

func cmdHelp() {
	fmt.Fprintln(os.Stderr, `Usage: getstocks [OPTION]...

getstocks retrives stocks quotes from web sites.
`)
}

func parseArgs() *clArgs {
	var args clArgs

	// command line arguments
	flag.StringVar(&args.config, clConfig, defaultConfigFile, "Configuration file used to generate the package.")
	flag.StringVar(&args.output, clOutput, defaultOutputFile, "Optional output file for package/config file. If empty stdout will be used.")
	flag.BoolVar(&args.debug, clDebug, false, "Debug mode. Do not cache templates and do not format generated code.")
	flag.BoolVar(&args.help, clHelp, false, "Show command usage information.")
	flag.BoolVar(&args.genConfig, clGenConfig, false, "Generate the configuration file instead of the package.")

	flag.Parse()

	return &args
}

// parseConfig returns the configuration from command line parameters,
// config parameters and defaults
func parseConfig(args *clArgs) (*config, error) {

	// init config from the config file
	cfg, err := loadConfigFromFile(args.config)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Run() int {
	const msghelp = "Try 'getstocks -h' for more information."

	args := parseArgs()

	if args.help {
		cmdHelp()
		return 0
	}

	// check config file exists
	if _, err := os.Stat(args.config); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Configuration file %q not found.\n%s\n", args.config, msghelp)
		return 2
	}

	// read config file
	cfg, err := parseConfig(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	fmt.Printf("cfg = %+v\n", cfg)

	return 0
}
