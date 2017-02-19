// Package cli provides a CLI UI for the getstocks command line tool.
package cli

import (
	"context"
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
	defaultConfigFile = "data-crypt/getstocks.cfg"
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
	Urls        []string            `toml:"urls"`
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
	setScrapers := run.NewSet()
	for _, scraper := range cfg.Scrapers {
		if len(scraper.Name) == 0 {
			return nil, errors.New("Invalid scraper: name must be defined")
		}
		if !setScrapers.Add(scraper.Name) {
			return nil, fmt.Errorf("Invalid scraper: name already used: %q", scraper.Name)
		}
	}

	// check stocks
	setStocks := run.NewSet()
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

func getRunArgs(cfg *config) ([]*run.Scraper, []*run.Stock, error) {
	disabledScrapers := run.NewSet()
	workers := map[string]int{}
	//usedScrapers := map[string]*run.Scraper{}

	stocks := make([]*run.Stock, 0, len(cfg.Stocks))

	for _, scr := range cfg.Scrapers {
		if scr.Disabled {
			disabledScrapers.Add(scr.Name)
			continue
		}
		workers[scr.Name] = scr.Workers
	}

	// Builds the stock array skipping only the esplicitly disabled stocks.
	// Stocks not disabled will be added even if without any valid source
	// in order to considered in the output result.
	for _, stock := range cfg.Stocks {
		if stock.Disabled {
			continue
		}

		// Builds array of StockSource from urls and sources.
		asrc := make([]*run.StockSource, 0, len(stock.Sources)+len(stock.Urls))

		// Checks stock.sources.
		for _, source := range stock.Sources {
			var err error
			if source.Disabled {
				continue
			}

			// Gets scraper name, if empty.
			if source.Scraper == "" {
				source.Scraper, err = run.GetScraperFromUrl(source.URL)
				//log.Printf("GetScraperFromUrl(%q) -> %s, %v", source.URL, source.Scraper, err)
			}
			// Skip if scraper is disabled.
			if disabledScrapers.Contains(source.Scraper) {
				continue
			}
			// Checks error eventually returned by GetScraperFromUrl.
			if err != nil {
				return nil, nil, err
			}
			// Append the new source to the list.
			asrc = append(asrc, &run.StockSource{
				Scraper: source.Scraper,
				URL:     source.URL,
			})
		}

		// Check stock.urls.
		for _, url := range stock.Urls {
			scraper, err := run.GetScraperFromUrl(url)
			// skip if scraper is disabled
			if disabledScrapers.Contains(scraper) {
				continue
			}
			// Checks error eventually returned by GetScraperFromUrl.
			if err != nil {
				return nil, nil, err
			}
			// append the new source to the list
			asrc = append(asrc, &run.StockSource{
				Scraper: scraper,
				URL:     url,
			})
		}
		// append the stock
		stocks = append(stocks, &run.Stock{
			Name:        stock.Name,
			Isin:        stock.Isin,
			Description: stock.Description,
			Sources:     asrc,
		})
	}

	// build scrapers array (only)
	scrapers := make([]*run.Scraper, 0, len(workers))
	for name, workers := range workers {
		scrapers = append(scrapers, &run.Scraper{
			Name:    name,
			Workers: workers,
		})
	}

	return scrapers, stocks, nil
}

func doJobOLD(scrapers []*run.Scraper, stocks []*run.Stock) error {

	ctx := context.Background()
	out, err := run.Execute(ctx, scrapers, stocks)

	if err != nil {
		return err
	}

	results := make([]*run.Response, 0, len(stocks))

	for r := range out {
		results = append(results, r)
	}

	for _, r := range results {
		var sPrice, sDate string

		if r.Err == nil {
			sPrice = fmt.Sprintf("%.3f", r.Price)
			sDate = r.Date.Format("02-01-2006")
		}
		fmt.Printf("%-20s %10s  %15s  (%s) %v\n", r.StockName, sPrice, sDate, r.ScraperName, r.Err)
	}
	return nil
}

func doJob(scrapers []*run.Scraper, stocks []*run.Stock) error {

	ctx := context.Background()
	out, err := run.Execute(ctx, scrapers, stocks)

	if err != nil {
		return err
	}

	fmt.Println("ISIN\tNAME\tPRICE\tDATE\tSCRAPER\tERROR")
	for r := range out {

		var sPrice, sDate string
		if r.Err == nil {
			sPrice = fmt.Sprintf("%.3f", r.Price)
			sDate = r.Date.Format("02-01-2006")
		}

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%v\n",
			"",
			r.StockName,
			sPrice,
			sDate,
			r.ScraperName,
			r.Err)
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
	//cfg.Print()

	// get scraper and stocks from config
	scrapers, stocks, err := getRunArgs(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	// call to run.Execute
	err = doJob(scrapers, stocks)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}
