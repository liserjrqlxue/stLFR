package main

import (
	"flag"
	"io"
	"log"
	"os"
	//"compress/gzip"
	gzip "github.com/klauspost/pgzip"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

var (
	stdout = flag.Bool(
		"c",
		false,
		"write on standard output, keep original files unchanged",
	)
	bufSize = flag.Int(
		"bufSize",
		1024*1024,
		"buffer size",
	)
)

func main() {
	flag.Parse()
	log.Printf("[%+v]\n", flag.Args())

	var buf = make([]byte, *bufSize)
	var files = flag.Args()

	var n int64

	if len(files) == 0 {
		var dest = os.Stdout
		var src = os.Stdin

		n = copy2gz(dest, src, buf)
	} else {
		for _, file := range files {
			if file == "-" {
				var dest = os.Stdout
				var src = os.Stdin

				n = copy2gz(dest, src, buf)

			} else {
				var dest io.WriteCloser
				if *stdout {
					dest = os.Stdout
				} else {
					dest = osUtil.Create(file + ".gz")
				}
				var src = osUtil.Open(file)

				n = copy2gz(dest, src, buf)
			}
		}
	}
	log.Printf("pgzip %d byte\n", n)
}

func copy2gz(dest io.WriteCloser, src io.ReadCloser, buf []byte) int64 {
	defer simpleUtil.DeferClose(src)
	defer simpleUtil.DeferClose(dest)

	var writer = gzip.NewWriter(dest)
	defer simpleUtil.DeferClose(writer)

	var n, err = io.CopyBuffer(writer, src, buf)

	if err != nil && err != io.EOF {
		log.Fatalf("Error after load %d bytes:[%+v]\n", n, err)
	}

	return n
}
