package main

import (
  "os"
  "bytes"
  "compress/flate"
  "io/ioutil"
  "flag"
  "fmt"
  "path/filepath"
  "strings"
  "net/http"
  "io"
)
const (
  FORMAT_TXT = "txt"
  FORMAT_PNG = "png"
  FORMAT_SVG = "svg"
  STYLE_TXT = "text"
  STYLE_LINK = "link"
  STYLE_OUTPUT = "output"
)
func main() {
  opts := parseArgs();
  if(opts.InputStream != nil){
    process(&opts, encodeAsTextFormat(opts.InputStream), "");
  }else{
    for _, filename := range opts.FileNames {
        data,err := ioutil.ReadFile(filename)
      if err != nil{
        fmt.Errorf("Error: Unable to read file %s\n", filename);
        os.Exit(1);
      }
	process(&opts,encodeAsTextFormat(data), filename);
    }
  }
}
func process(options *Options, textFormat string, filename string) {
  if options.Style == STYLE_TXT {
    fmt.Printf("%s\n" , textFormat)
  } else if options.Style == STYLE_LINK {
    fmt.Printf("%s/%s/%s\n", options.Server, options.Format, textFormat)
  } else if options.Style == STYLE_OUTPUT {
    link := fmt.Sprintf("%s/%s/%s", options.Server, options.Format, textFormat)
    var output *os.File;

    if filename == "" {
	output = os.Stdout
    }else{
      outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename))
      outputFilename = fmt.Sprintf("%s.%s", outputFilename, options.Format)
      output, _ = os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE, 0666)
    }
    response, err := http.Get(link)
    if err != nil {
	fmt.Errorf("Error Fetching: %s\n", link)
        return
    }
    if response.StatusCode != 200 {
      fmt.Errorf("Error Fetch: %s\n%s\n", link, response.Status)
      return
    }

    io.Copy(output, response.Body)
    output.Close()
  }
}

type Options struct {
  Server string
  Format string
  Style string
  InputStream []byte
  FileNames []string
}

func encodeAsTextFormat(input []byte) string{
  return encode64(deflate(input));
}

func deflate(input []byte) []byte {
  var b bytes.Buffer
  w, _ := flate.NewWriter(&b, flate.BestCompression)
  w.Write(input)
  w.Close()
  return b.Bytes()
}

func encode64(input []byte) string {

  var part []byte
  var buffer bytes.Buffer

  inputLength := len(input)

  for i := 0; i < inputLength; i += 3 {
    if i+2 == inputLength {
      part = to4Bytes(input[i], input[i+1], 0)
    } else if i+1 == inputLength {
      part = to4Bytes(input[i], 0, 0)
    } else {
      part = to4Bytes(input[i], input[i+1], input[i+2])
    }
    buffer.Write(part);
  }

  return string(buffer.Bytes())
}
func to4Bytes(b1, b2, b3 byte) []byte {
  c1 := b1 >> 2
  c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
  c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
  c4 := b3 & 0x3F
  return []byte{
    encode6bit(c1 & 0x3F),
    encode6bit(c2 & 0x3F),
    encode6bit(c3 & 0x3F),
    encode6bit(c4 & 0x3F),
  }
}

func encode6bit(b byte) byte {
  if b < 10 {
    return byte(48 + b)
  }
  b -= 10
  if b < 26 {
    return byte(65 + b)
  }
  b -= 26
  if b < 26 {
    return byte(97 + b)
  }
  b -= 26
  if b == 0 {
    return ([]byte("-"))[0]
  }
  if b == 1 {
    return ([]byte("_"))[0]
  }
  return ([]byte("?"))[0]
}


func parseArgs() Options{
  flag.CommandLine.Init(os.Args[0],flag.ExitOnError)
  server := flag.String("s", "http://plantuml.com/plantuml", "Plantuml `server` address. Used when generating link or extracting output")
  format := flag.String("f", "png", "Output `format` type. (Options: png,txt,svg)")
  style := flag.String("o", "text", "Indicates if `output` style. (Options: text, link, output)")
  help := flag.Bool("h", false, "Show help (this) text")
  flag.Parse()
  files := flag.Args()

  var inputStream []byte

  stat, _ := os.Stdin.Stat()
  if (stat.Mode() & os.ModeCharDevice) == 0 {
    data,err := ioutil.ReadAll(os.Stdin)
    if err == nil {
      inputStream = data
    }
  }
  if *help || len(files) == 0 && len(inputStream) == 0  {
    fmt.Printf(`USAGE:
    plantuml-go [OPTIONS] files
        Reads and process files based on options
    plantuml-go [OPTIONS]
        Reads and process stdin. NOTE: Ouput will be on stdout
OPTIONS
`,);
    flag.PrintDefaults()
    os.Exit(1)
  }
  opts := Options{*server, *format,*style, inputStream, files}
  return opts;
}