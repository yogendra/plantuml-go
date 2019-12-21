package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	formatTxt   = "txt"
	formatPng   = "png"
	formatSvg   = "svg"
	styleTxt    = "text"
	styleLink   = "link"
	styleOutput = "output"
	mapper      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
)

func main() {
	opts, err := parseArgs()
	if err != nil {
		log.Println("failed to parse args:", err)
		return
	}

	if opts.InputStream != nil {
		process(&opts, encodeAsTextFormat(opts.InputStream), "")
	} else {
		for _, filename := range opts.FileNames {
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			process(&opts, encodeAsTextFormat(data), filename)
		}
	}
}

func process(options *options, textFormat string, filename string) {
	if options.Style == styleTxt {
		fmt.Printf("%s\n", textFormat)
	} else if options.Style == styleLink {
		u, err := url.Parse(options.Server)
		if err != nil {
			fmt.Println("failed to parse the url:", options.Server)
			return
		}
		u.Path = path.Join(u.Path, options.Format, textFormat)
		fmt.Println(u.String())
	} else if options.Style == styleOutput {
		u, err := url.Parse(options.Server)
		if err != nil {
			fmt.Println("failed to parse the url:", options.Server)
			return
		}
		u.Path = path.Join(u.Path, options.Format, textFormat)
		link := u.String()
		output := os.Stdout

		if filename != "" {
			outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename))
			outputFilename = fmt.Sprintf("%s.%s", outputFilename, options.Format)
			output, _ = os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE, 0666)
		}
		response, err := http.Get(link)
		if err != nil {
			fmt.Println(err)
			return
		}
		if response.StatusCode != 200 {
			fmt.Printf("Error in Fetching: %s\n%s\n", link, response.Status)
			return
		}

		io.Copy(output, response.Body)
		output.Close()
	}
}

type options struct {
	Server      string
	Format      string
	Style       string
	InputStream []byte
	FileNames   []string
}

func encodeAsTextFormat(raw []byte) string {
	compressed := deflate(raw)
	return base64Encode(compressed)
}

func deflate(input []byte) []byte {
	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, zlib.BestCompression)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func base64Encode(input []byte) string {
	var buffer bytes.Buffer
	inputLength := len(input)
	for i := 0; i < 3-inputLength%3; i++ {
		input = append(input, byte(0))
	}

	for i := 0; i < inputLength; i += 3 {
		b1, b2, b3, b4 := input[i], input[i+1], input[i+2], byte(0)

		b4 = b3 & 0x3f
		b3 = ((b2 & 0xf) << 2) | (b3 >> 6)
		b2 = ((b1 & 0x3) << 4) | (b2 >> 4)
		b1 = b1 >> 2

		for _, b := range []byte{b1, b2, b3, b4} {
			buffer.WriteByte(byte(mapper[b]))
		}
	}
	return string(buffer.Bytes())
}

func parseArgs() (options, error) {
	flag.CommandLine.Init(os.Args[0], flag.ExitOnError)
	server := flag.String("s", "http://plantuml.com/plantuml", "Plantuml `server` address. Used when generating link or extracting output")
	format := flag.String("f", "png", "Output `format` type. (Options: png,txt,svg)")
	style := flag.String("o", "text", "Indicates if `output` style. (Options: text, link, output)")
	help := flag.Bool("h", false, "Show help (this) text")
	flag.Parse()

	fileList := flag.Args()
	files := []string{}
	if len(fileList) > 0 {
		fileMap := make(map[string]int)
		for _, f := range fileList {
			abs, err := filepath.Abs(f)
			if err != nil {
				return options{}, fmt.Errorf("%s: unable to resolve filename", f)
			}
			_, ok := fileMap[abs]
			if !ok {
				files = append(files, abs)
				fileMap[abs] = 1
			}
		}
	}
	var inputStream []byte
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := ioutil.ReadAll(os.Stdin)
		if err == nil {
			inputStream = data
		}
	}

	if *help || len(files) == 0 && len(inputStream) == 0 {
		fmt.Printf(`USAGE:
    plantuml-go [OPTIONS] files
        Reads and process files based on options
    plantuml-go [OPTIONS]
        Reads and process stdin. NOTE: Ouput will be on stdout
OPTIONS
`)
		flag.PrintDefaults()
		os.Exit(1)
	}
	opts := options{*server, *format, *style, inputStream, files}
	return opts, nil
}
