package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	//"compress/gzip"
	gzip "github.com/klauspost/pgzip"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "..", "BIN")
)

var (
	barcodeList = flag.String(
		"bc",
		filepath.Join(dbPath, "barcode.list"),
		"barcode list",
	)
	read1 = flag.String(
		"fq1",
		"",
		"read 1 of PE",
	)
	read2 = flag.String(
		"fq2",
		"",
		"read 2 of PE",
	)
	readLength = flag.Int(
		"l",
		100,
		"read length",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output[-prefix.{1,2}.fq.gz",
	)
)

func main() {
	flag.Parse()

	// load barcode list
	// add 1 mismatch
	var (
		barcodeHash = make(map[string]string)
		n           = 0
	)
	var bl = osUtil.Open(*barcodeList)
	defer simpleUtil.DeferClose(bl)
	var blScan = bufio.NewScanner(bl)
	for blScan.Scan() {
		var line = strings.Split(blScan.Text(), "\t")
		n++
		var barcodeId = line[1]
		for _, r := range []rune("ACGT") {
			for i := 0; i < 10; i++ {
				var barcodeMis = []rune(line[0])
				barcodeMis[i] = r
				barcodeHash[string(barcodeMis)] = barcodeId
			}
		}
	}

	var fq1 = osUtil.Open(*read1)
	defer simpleUtil.DeferClose(fq1)
	var fq2 = osUtil.Open(*read2)
	defer simpleUtil.DeferClose(fq2)

	var fq1Gr = simpleUtil.HandleError(gzip.NewReader(fq1)).(*gzip.Reader)
	var fq1Scanner = bufio.NewScanner(fq1Gr)
	simpleUtil.DeferClose(fq1Gr)
	var fq2Gr = simpleUtil.HandleError(gzip.NewReader(fq2)).(*gzip.Reader)
	var fq2Scanner = bufio.NewScanner(fq2Gr)
	simpleUtil.DeferClose(fq2Gr)

	var outFq1 = osUtil.Create(filepath.Join(*prefix, ".1.fq.gz"))
	defer simpleUtil.DeferClose(outFq1)
	var outFq2 = osUtil.Create(filepath.Join(*prefix, ".2.fq.gz"))
	defer simpleUtil.DeferClose(outFq2)

	var outFq1Zw = gzip.NewWriter(outFq1)
	defer simpleUtil.DeferClose(outFq1Zw)
	var outFq2Zw = gzip.NewWriter(outFq2)
	defer simpleUtil.DeferClose(outFq2Zw)

	var (
		// barcode seq
		// ACGTACGTAC NNNNNN GTACGTACGT NNNNNN ACGTACGTAC
		// n1         n2     n3         n4     n5
		n1, n2, n3, n4, n5 = 10, 6, 10, 6, 10
		length1            = *readLength
		length1e           = length1 + n1
		length2            = *readLength + n1 + n2
		length2e           = length2 + n3
		length3            = *readLength + n1 + n2 + n3 + n4
		length3e           = length3 + n5
		readCount          = 0
		splitCount         = 0
		splitBarcodeNum    = 0
		splitBarcodeHash   = make(map[string]int)
		read1Name          string
		read2Name          string
	)
	for fq1Scanner.Scan() {
		if !fq2Scanner.Scan() {
			log.Fatal("Pair End Error!")
		}
		readCount++
		switch readCount % 4 {
		case 1:
			read1Name = strings.Split(fq1Scanner.Text(), "/")[0]
			read2Name = strings.Split(fq2Scanner.Text(), "/")[0]
			if read1Name != read2Name {
				log.Fatalf("Error: [%s] not eq [%s] at the %d reads", read1Name, read2Name, readCount)
			}
		case 2:
			var line1 = fq1Scanner.Bytes()
			var line2 = fq1Scanner.Bytes()
			var (
				b1id, ok1  = barcodeHash[string(line2[length1:length1e])]
				b2id, ok2  = barcodeHash[string(line2[length2:length2e])]
				b3id, ok3  = barcodeHash[string(line2[length3:length3e])]
				barcode    = "0_0_0"
				barcodeNum = 0
			)
			if ok1 && ok2 && ok3 {
				barcode = b1id + "_" + b2id + "_" + b3id
				splitCount++
				if splitBarcodeHash[barcode] == 0 {
					splitBarcodeNum++
					splitBarcodeHash[barcode] = splitBarcodeNum
				}
				barcodeNum = splitBarcodeHash[barcode]
			}
			simpleUtil.HandleError(
				outFq1Zw.Write(
					[]byte(fmt.Sprintf("%s#%s/1\t%d\t1\n%s\n", read1Name, barcode, barcodeNum, line1[0:length1])),
				),
			)
			simpleUtil.HandleError(
				outFq2Zw.Write(
					[]byte(fmt.Sprintf("%s#%s/2\t%d\t1\n%s\n", read2Name, barcode, barcodeNum, line2[0:length1])),
				),
			)
		case 3:
			simpleUtil.HandleError(
				outFq1Zw.Write(
					append(fq1Scanner.Bytes(), '\n'),
				),
			)
			simpleUtil.HandleError(
				outFq2Zw.Write(
					append(fq2Scanner.Bytes(), '\n'),
				),
			)
		case 0:
			simpleUtil.HandleError(
				outFq1Zw.Write(
					append(fq1Scanner.Bytes()[0:length1], '\n'),
				),
			)
			simpleUtil.HandleError(
				outFq2Zw.Write(
					append(fq2Scanner.Bytes()[0:length1], '\n'),
				),
			)
		}
	}

	var (
		barcodeTypes = n * n * n
	)
	fmt.Printf("Barcode_types = %d * %d * %d = %d\n", n, n, n, barcodeTypes)
	fmt.Printf(
		"Real_Barcode_types = %d (%f %%)\n",
		splitBarcodeNum, float64(splitBarcodeNum)/float64(barcodeTypes)*100,
	)
	fmt.Printf("Reads_pair_num = %d\n", readCount/4)
	fmt.Printf("Reads_pair_num(after split) = %d (%f %%)\n", splitCount, float64(splitCount)/float64(readCount)*100)
	if readCount%4 != 0 {
		log.Fatalf("Error: fastq line error:[%s]", readCount)
	}
}
