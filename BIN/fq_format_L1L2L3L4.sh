#!/bin/bash
date
read1_L1=$1
read2_L1=$2
read1_L2=$3
read2_L2=$4
read1_L3=$5
read2_L3=$6
read1_L4=$7
read2_L4=$8

id=$9
if [ ! -d "$id" ]
then
	`mkdir -p $id`
fi

bin=/zfstj1/B2C_RD_P2/BC_RDC1/USER/zhuyaping/work/stLFR/BIN_NEW
cat ${read1_L1} ${read1_L2} ${read1_L3} ${read1_L4} >${id}/${id}_read_1.fq.gz 
cat ${read2_L1} ${read2_L2} ${read2_L3} ${read2_L4} >${id}/${id}_read_2.fq.gz
date

perl ${bin}/split_barcode_PEXXX_42_reads_zyp.pl ${bin}/barcode.list ${bin}/barcode_RC.list ${id}/${id}_read_1.fq.gz ${id}/${id}_read_2.fq.gz 100 ${id}/${id}_split_read 
#perl ${bin}/split_barcode_PEXXX_42_reads.2.pl ${bin}/barcode.list ${bin}/barcode_RC.list ${id}/${id}_read_1.fq.gz ${id}/${id}_read_2.fq.gz 100 ${id}/${id}_split_read 
date

perl -e '$n=0;while(<>){$n++;chomp;@t=split; if($n>=5){$B=$t[1]*(100+100); $G=$B/(1e+6); print "$t[0]\t$t[1]\t$B\t$G\t$t[2]\n"; } }' ${id}/${id}_split_read_split_stat_read1.log >${id}/${id}_BaseSize.stat
awk '{print $4}' ${id}/${id}_BaseSize.stat >${id}/${id}_BaseSize.stat.plot
perl ${bin}/transfer_modified_v3.pl ${id}/${id}_split_read.1.fq.gz ${id}/${id}_split_read.2.fq.gz ${bin}/4M-with-alts-february-2016.txt ${id}/${id}
date
echo done!
