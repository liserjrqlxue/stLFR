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
	barcodeRCList = flag.String(
		"rc",
		filepath.Join(dbPath, "barcode_RC.list"),
		"barcod RC list",
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
	var (
		barcode_hash = make(map[string]string)
		n            = 0
	)
	var bl = osUtil.Open(*barcodeList)
	defer simpleUtil.DeferClose(bl)
	var blScan = bufio.NewScanner(bl)
	for blScan.Scan() {
		var line = strings.Split(blScan.Text(), "\t")
		n++
		var barcode_ID = line[1]
		for _, r := range []rune("ACGT") {
			for i := 0; i < 10; i++ {
				var barcode_mis = []rune(line[0])
				barcode_mis[i] = r
				barcode_hash[string(barcode_mis)] = barcode_ID
			}
		}
	}

	var fq1 = osUtil.Open(*read1)
	defer simpleUtil.DeferClose(fq1)
	var fq2 = osUtil.Open(*read1)
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
		n1, n2, n3, n4, n5 = 10, 6, 10, 6, 10
		length1            = *readLength
		length1e           = length1 + n1
		length2            = *readLength + n1 + n2
		length2e           = length2 + n3
		length3            = *readLength + n1 + n2 + n3 + n4
		length3e           = length3 + n5
		barcode_types      = n * n * n
		num1               = 0
		num2               = 0
		k1, k2             string
		hash               = make(map[string]int)
		split_barcode_num  = 0
	)
	for fq1Scanner.Scan() {
		if !fq2Scanner.Scan() {
			log.Fatal("Pair End Error!")
		}
		num1++
		switch num1 % 4 {
		case 1:
			k1 = strings.Split(fq1Scanner.Text(), "/")[0]
			k2 = strings.Split(fq2Scanner.Text(), "/")[0]
			if k1 != k2 {
				log.Fatalf("Error: [%s] not eq [%s] at the %d reads", k1, k2, num1)
			}
		case 2:
			var line1 = fq1Scanner.Bytes()
			var line2 = fq1Scanner.Bytes()
			var (
				b1id, ok1 = barcode_hash[string(line2[length1:length1e])]
				b2id, ok2 = barcode_hash[string(line2[length2:length2e])]
				b3id, ok3 = barcode_hash[string(line2[length3:length3e])]
				id        = "0_0_0"
				hash_id   = 0
			)
			if ok1 && ok2 && ok3 {
				id = b1id + "_" + b2id + "_" + b3id
				num2++
				if hash[id] == 0 {
					split_barcode_num++
					hash[id] = split_barcode_num
				}
				hash_id = hash[id]
			}
			simpleUtil.HandleError(
				outFq1Zw.Write(
					[]byte(fmt.Sprintf("%s#%s/1\t%d\t1\n%s\n", k1, id, hash_id, line1[0:length1])),
				),
			)
			simpleUtil.HandleError(
				outFq2Zw.Write(
					[]byte(fmt.Sprintf("%s#%s/2\t%d\t1\n%s\n", k2, id, hash_id, line2[0:length1])),
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

	fmt.Printf("Barcode_types = %d * %d * %d = %d\n", n, n, n, barcode_types)
	fmt.Printf(
		"Real_Barcode_types = %d (%f %%)\n",
		split_barcode_num, float64(split_barcode_num)/float64(barcode_types)*100,
	)
	fmt.Printf("Reads_pair_num = %d\n", num1/4)
	fmt.Printf("Reads_pair_num(after split) = %d (%f %%)\n", num2, float64(num2)/float64(num1)*100)
	if num1%4 != 0 {
		log.Fatalf("Error: fastq line error:[%s]", num1)
	}
}
