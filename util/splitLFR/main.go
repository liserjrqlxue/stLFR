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
	dbPath = filepath.Join(exPath, "..", "..", "db")
)

var (
	barcodeList = flag.String(
		"bc",
		filepath.Join(dbPath, "barcode.list"),
		"barcode list",
	)
	mapList = flag.String(
		"map",
		filepath.Join(dbPath, "4M-with-alts-february-2016.txt.gz"),
		"map list",
	)
	read1 = flag.String(
		"fq1",
		"",
		"read 1 of PEs, comma as sep",
	)
	read2 = flag.String(
		"fq2",
		"",
		"read 2 of PEs, comma as sep, same order with -fq1",
	)
	readLength = flag.Int(
		"l",
		100,
		"read length",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output[-prefix_split_read.{1,2}.fq.gz",
	)
)

func main() {
	flag.Parse()
	if *read1 == "" || *read2 == "" || *prefix == "" {
		flag.Usage()
		log.Fatal("-fq1/-fq2/-prefix required!")
	}

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

	// load map list
	var ml = osUtil.Open(*mapList)
	defer simpleUtil.DeferClose(ml)
	var mlGr = simpleUtil.HandleError(gzip.NewReader(ml)).(*gzip.Reader)
	defer simpleUtil.DeferClose(mlGr)
	var mlScan = bufio.NewScanner(mlGr)
	var mapHash = make(map[int]string)
	var mlCount = 0
	for mlScan.Scan() {
		mlCount++
		mapHash[mlCount] = mlScan.Text()
	}

	var reads1 = strings.Split(*read1, ",")
	var reads2 = strings.Split(*read2, ",")
	if len(reads1) != len(reads2) {
		log.Fatalf("Error: incompatible of pair end")
	}
	var pairEnds [][2]string
	for i := range reads1 {
		pairEnds = append(pairEnds, [2]string{reads1[i], reads2[i]})

	}

	simpleUtil.CheckErr(os.MkdirAll(filepath.Dir(*prefix), 0700))
	var outFq1 = osUtil.Create(*prefix + "_split_read.1.fq.gz")
	defer simpleUtil.DeferClose(outFq1)
	var outFq2 = osUtil.Create(*prefix + "_split_read.2.fq.gz")
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

	for _, pairEnd := range pairEnds {
		var fq1 = osUtil.Open(pairEnd[0])
		var fq2 = osUtil.Open(pairEnd[1])

		var fq1Gr = simpleUtil.HandleError(gzip.NewReader(fq1)).(*gzip.Reader)
		var fq1Scanner = bufio.NewScanner(fq1Gr)
		var fq2Gr = simpleUtil.HandleError(gzip.NewReader(fq2)).(*gzip.Reader)
		var fq2Scanner = bufio.NewScanner(fq2Gr)

		for fq1Scanner.Scan() {
			if !fq2Scanner.Scan() {
				log.Fatal("Pair End Error!")
			}
			readCount++
			switch readCount % 4 {
			case 1: // name
				read1Name = strings.Split(fq1Scanner.Text(), "/")[0]
				read2Name = strings.Split(fq2Scanner.Text(), "/")[0]
				if read1Name != read2Name {
					log.Fatalf("Error: [%s] not eq [%s] at the %d reads", read1Name, read2Name, readCount)
				}
			case 2: // seq
				var (
					line1      = fq1Scanner.Bytes()
					line2      = fq2Scanner.Bytes()
					seq1       = line1[0:length1]
					seq2       = line2[0:length1]
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
						[]byte(
							fmt.Sprintf(
								"%s#%s/1\t%d\t1\n%s\n",
								read1Name, barcode, barcodeNum, seq1,
							),
						),
					),
				)
				simpleUtil.HandleError(
					outFq2Zw.Write(
						[]byte(
							fmt.Sprintf(
								"%s#%s/2\t%d\t1\n%s\n",
								read2Name, barcode, barcodeNum, seq2),
						),
					),
				)
			case 3: // note
				var line1 = append(fq1Scanner.Bytes(), '\n')
				var line2 = append(fq2Scanner.Bytes(), '\n')
				simpleUtil.HandleError(outFq1Zw.Write(line1))
				simpleUtil.HandleError(outFq2Zw.Write(line2))
			case 0: // qual
				var line1 = append(fq1Scanner.Bytes()[0:length1], '\n')
				var line2 = append(fq2Scanner.Bytes()[0:length1], '\n')
				simpleUtil.HandleError(outFq1Zw.Write(line1))
				simpleUtil.HandleError(outFq2Zw.Write(line2))
			}
		}

		simpleUtil.CheckErr(fq1Gr.Close())
		simpleUtil.CheckErr(fq2Gr.Close())
		simpleUtil.CheckErr(fq1.Close())
		simpleUtil.CheckErr(fq2.Close())
	}

	var (
		barcodeTypes = n * n * n
	)
	fmt.Printf("Barcode_types = %d * %d * %d = %d\n", n, n, n, barcodeTypes)
	fmt.Printf(
		"Real_Barcode_types = %d (%f %%)\n",
		splitBarcodeNum, float64(splitBarcodeNum)/float64(barcodeTypes)*100,
	)
	if readCount%4 != 0 {
		log.Fatalf("Error: fastq line error:[%s]", readCount)
	}
	readCount /= 4
	fmt.Printf("Reads_pair_num = %d\n", readCount)
	fmt.Printf("Reads_pair_num(after split) = %d (%f %%)\n", splitCount, float64(splitCount)/float64(readCount)*100)
}
