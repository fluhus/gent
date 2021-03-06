// Command goat generates go source from a given template.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"text/template"
)

var (
	in   = flag.String("i", "", "Path to input template file. If omitted, reads from stdin.")
	out  = flag.String("o", "", "Path to output go file. If omitted, writes to stdout.")
	data = flag.String("d", "", "JSON-encoded data for the template.")
	nh   = flag.Bool("nh", false, "Don't add a header to the output file.")
	nf   = flag.Bool("nf", false, "Don't run gofmt on the result.")
)

func main() {
	flag.Parse()

	// Parse template data.
	var d interface{}
	if *data != "" {
		if err := json.Unmarshal([]byte(*data), &d); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse data (-d param):", err)
			os.Exit(2)
		}
	}

	// Read template.
	var err error
	var input []byte
	if *in == "" {
		fmt.Fprintln(os.Stderr, "Reading from stdin...")
		input, err = ioutil.ReadAll(os.Stdin)
	} else {
		input, err = ioutil.ReadFile(*in)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read input:", err)
		os.Exit(2)
	}

	// Parse template.
	funcs := map[string]interface{}{"slice": makeSlice}
	t, err := template.New("").Funcs(funcs).Parse(string(input))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse template:", err)
		os.Exit(2)
	}

	// Execute template.
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, d)
	if err != nil {
		fmt.Println("Failed to execute template:", err)
		os.Exit(2)
	}
	src := buf.Bytes()

	// Attach header.
	if !*nh {
		from := ""
		if *in != "" {
			from = "from '" + *in + "' "
		}
		header = fmt.Sprintf(header, from)
		src = append([]byte(header), src...)
	}

	// Run gofmt.
	if !*nf {
		src, err = format.Source(src)
		if err != nil {
			fmt.Println("Failed to gofmt the resulting source:", err)
			os.Exit(2)
		}
	}

	// Write output.
	if *out == "" {
		fmt.Print(string(src))
	} else {
		err = ioutil.WriteFile(*out, src, 0644)
		if err != nil {
			fmt.Println("Failed to write output:", err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "Wrote to:", *out)
	}
}

func makeSlice(a ...interface{}) []interface{} {
	return a
}

var header = `// ***** DO NOT EDIT THIS FILE MANUALLY. *****
//
// This file was auto-generated %vusing goat.
//
// goat: https://www.github.com/fluhus/goat

`
