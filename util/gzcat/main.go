package main

import (
	"flag"
	"io"
	"log"
	"regexp"
	"strings"

	//"compress/gzip"
	gzip "github.com/klauspost/pgzip"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

var (
	isGz = regexp.MustCompile(`.gz$`)
)

var (
	input = flag.String(
		"input",
		"",
		"gz file list, comma as sep",
	)
	output = flag.String(
		"out",
		"",
		"cat output, .gz suffix support",
	)
)

func main() {
	flag.Parse()
	if *input == "" || *output == "" {
		flag.Usage()
		log.Fatal("-input/output required!")
	}

	var outF = osUtil.Create(*output)
	defer simpleUtil.DeferClose(outF)
	var outZw *gzip.Writer
	if isGz.MatchString(*output) {
		outZw = gzip.NewWriter(outF)
		defer simpleUtil.DeferClose(outZw)
	}

	var n1, n2 int
	for _, inputFile := range strings.Split(*input, ",") {
		var inF = osUtil.Open(inputFile)
		var inZr *gzip.Reader
		if isGz.MatchString(inputFile) {
			inZr = simpleUtil.HandleError(gzip.NewReader(inF)).(*gzip.Reader)
		}

		for {
			var buf = make([]byte, 1024*1024)
			var n, err = inZr.Read(buf)
			if err == io.EOF {
				break
			}
			simpleUtil.CheckErr(err)
			n1 += n
			n, err = outZw.Write(buf[:n])
			simpleUtil.CheckErr(err)
			n2 += n
		}
		if inZr != nil {
			simpleUtil.CheckErr(inZr.Close())
		}
		simpleUtil.CheckErr(inF.Close())
	}

	if outZw != nil {
		simpleUtil.CheckErr(outZw.Close())
	}
	log.Printf("load %d byte,write %d byte\n", n1, n2)
}
