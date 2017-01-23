// Package cli provides a CLI UI for the gentmpl command line tool.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mmbros/getstocks/run"
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
	Name     string
	Workers  int
	Disabled bool
}

type configStock struct {
	Name        string
	Isin        string
	Description string
	Disabled    bool
	Sources     []configStockSource `toml:"source"`
}

type configStockSource struct {
	Scraper  string
	URL      string
	Disabled bool
}

type config struct {
	Scrapers []*configScraper `toml:"scraper"`
	Stocks   []*configStock   `toml:"stock"`
}

func (cfg *config) Print() {
	for j, s := range cfg.Scrapers {
		fmt.Printf("Scrapers[%d] %v\n", j, s)
	}
	for j, s := range cfg.Stocks {
		fmt.Printf("\nStocks[%d] %v\n", j, s)
	}
}

func unmarshalConfig(data []byte) (*config, error) {
	// parse config file
	var cfg config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// check scrapers
	setScrapers := NewSet()
	for _, scraper := range cfg.Scrapers {
		if len(scraper.Name) == 0 {
			return nil, errors.New("Invalid scraper: name must be defined")
		}
		if !setScrapers.Add(scraper.Name) {
			return nil, fmt.Errorf("Invalid scraper: name already used: %q", scraper.Name)
		}
	}

	// check stocks
	setStocks := NewSet()
	for _, stock := range cfg.Stocks {
		if len(stock.Name) == 0 {
			return nil, errors.New("Invalid stock: name must be defined")
		}
		if !setStocks.Add(stock.Name) {
			return nil, fmt.Errorf("Invalid stock: name already used: %q", stock.Name)
		}

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

func getRunArgs(cfg *config) error {
	disabledScrapers := NewSet()
	workers := map[string]int{}
	usedScrapers := map[string]*run.Scraper{}

	for _, scr := range cfg.Scrapers {
		if scr.Disabled {
			disabledScrapers.Add(scr.Name)
			continue
		}
		workers[scr.Name] = scr.Workers
	}

	for _, stock := range cfg.Stocks {
		if stock.Disabled {
			continue
		}
		for _, source := range stock.Sources {
			if source.Disabled {
				continue
			}

			// get the scraper type
			scrapertype, err := run.ScraperTypeFromStringOrUrl(source.Scraper, source.URL)
			if err != nil {
				return err
			}
			// check if scraper name is disabled
			scrapername := scrapertype.String()
			if disabledScrapers.Contains(scrapername) {
				continue
			}
			// add the scraper to the used scrapers, if not already in use
			if _, ok := usedScrapers[scrapername]; !ok {
				runscr, err := run.NewScraper(scrapertype, workers[scrapername])
				if err != nil {
					return err
				}
				usedScrapers[scrapername] = runscr
			}
		}
	}

	fmt.Println("------------------")
	for name, scr := range usedScrapers {
		fmt.Printf("%s -> %d\n", name, scr.Workers())
	}

	return nil
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
	cfg.Print()

	err = getRunArgs(cfg)
	if err != nil {
		fmt.Println("XXXXXXXXX ", err)
	}

	return 0
}
