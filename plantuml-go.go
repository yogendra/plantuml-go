package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	styleTxt    = "text"
	styleLink   = "link"
	styleOutput = "output"
	mapper      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
)

type option struct {
	server string
	format string
	style  string
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

// getImage from the PlantUML Server with the url and writes the
// image data to the w writer.
func getImage(url string, w io.Writer) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get image from %s:%s", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch because the status code was %v", res.StatusCode)
	}

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return fmt.Errorf("failed to copy the image to the writer:%s", err)
	}

	return nil
}

func getImageWithFileList(opt option, list []string) error {
	var e error
	for _, f := range list {
		abs, err := filepath.Abs(f)
		if err != nil {
			e = fmt.Errorf("%s: failed to get the absolute file path: %s", e, f)
		}
		out := strings.TrimSuffix(abs, filepath.Ext(abs))
		out = fmt.Sprintf("%s.%s", out, opt.format)

		output, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			e = fmt.Errorf("%s: failed to open the file: %s", e, err)
		}
		defer output.Close()

		data, err := ioutil.ReadFile(abs)
		if err != nil {
			e = fmt.Errorf("%s: failed to read the file: %s", e, err)
		}

		err = getImageWithOneStream(opt, data, output)
		if err != nil {
			e = fmt.Errorf("%s: %s", e, err)
		}
	}

	return e
}

func getImageWithOneStream(opt option, data []byte, w io.Writer) error {
	encorded := encodeAsTextFormat(data)

	u, err := url.Parse(opt.server)
	if err != nil {
		fmt.Printf("failed to parse the url '%s': %s\n", opt.server, err)
	}
	u.Path = path.Join(u.Path, opt.format, encorded)
	link := u.String()

	switch opt.style {
	case styleTxt:
		fmt.Println(encorded)
	case styleLink:
		fmt.Println(link)
	case styleOutput:
		err := getImage(link, w)
		if err != nil {
			return fmt.Errorf("failed to get image from %s:%s", link, err)
		}
	default:
		return fmt.Errorf("style '%s' is invalid", opt.style)
	}
	return nil
}
