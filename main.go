package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	usage = `USAGE:
    plantuml-go [OPTIONS] files
        Reads and process files based on options
    plantuml-go [OPTIONS]
        Reads and process stdin. NOTE: Ouput will be on stdout
OPTIONS
`
)

var (
	opt      option
	help     bool
	fileList []string
)

func init() {
	flag.StringVar(&opt.server, "s", "http://plantuml.com/plantuml", "Plantuml `server` address. Used when generating link or extracting output")
	flag.StringVar(&opt.format, "f", "png", "Output `format` type. (Options: png,txt,svg)")
	flag.StringVar(&opt.style, "o", "text", "Indicates if `output` style. (Options: text, link, output)")
	flag.BoolVar(&help, "h", false, "Show help (this) text")
	flag.Parse()
}

func main() {
	fileList = flag.Args()
	os.Exit(run())
}

func run() int {
	if help {
		fmt.Println(usage)
		flag.PrintDefaults()
		return 0
	}

	inputStream, err := getInputStream()
	if err != nil {
		fmt.Println("failed to get input stream:", err)
		return 1
	}

	if len(inputStream) == 0 {
		if len(fileList) == 0 {
			fmt.Println("Please piped PlantUML data or specify puml files.")
			fmt.Println(usage)
			flag.PrintDefaults()
			return 1
		}

		err = getImageWithFileList(opt, fileList)
		if err != nil {
			fmt.Println(err)
			return 1
		}
	} else {
		err = getImageWithOneStream(opt, inputStream, os.Stdout)
		if err != nil {
			fmt.Println(err)
			return 1
		}
	}

	return 0
}

func getInputStream() ([]byte, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get the stat of the Stdin:%s", err)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}
	return nil, nil
}
