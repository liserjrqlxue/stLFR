#!/usr/bin/perl

#####script from Guo Lidong at BGI, 20180702, modified to print read1 and read2 into seperate files, and named as read 1 and 2, also indexes with names '1' instead of '2' #######
###add file index so that when processing split fastq from same library, reads won't have the same name

###add 16bp barcode and 7bp seq before read1, add sample index at the end of read names
print "Merge stLFR reads into 10X format !\n read1 :  $ARGV[0] \n. read2 : $ARGV[1] \n map file : $ARGV[2] \n prefix: $ARGV[3] \n" ;
open IN3,$ARGV[2] ;
my $n=0;
$barcode_num=1;
while(<IN3>)
{
	chomp;
	$n++;
    my @pair =split(/\t/,$_);
    $map{$n}=$pair[0] ;
}
close IN3;

open IN1,"gzip -dc $ARGV[0] | ";
open IN2,"gzip -dc $ARGV[1] | ";
my $id=$ARGV[3];
#open OUT,"| gzip > read-R1_si-TTCACGCG_lane-001-chunk-001.fastq.gz"; 
#open OUT2,"| gzip > read-I1_si-TTCACGCG_lane-001-chunk-001.fastq.gz";  
#open OUT3,"| gzip > read-R2_si-TTCACGCG_lane-001-chunk-001.fastq.gz";
open OUT,"| gzip > $id\_S1_L001_R1_001.fastq.gz";
open OUT3,"| gzip > $id\_S1_L001_R2_001.fastq.gz";

$N=0;
while(<IN1>)
{
    chomp;
    my @name=split(/\t/, $_);
	my $temp=$name[1] % $n;
	if ($temp==0){
		$temp=$n;
	}
    if(! exists($map{$temp}) )
    {
        $S=<IN1>;
        $S=<IN1>;
        $S=<IN1>;

        $S=<IN2>;
        $S=<IN2>;
        $S=<IN2>;
        $S=<IN2>;

        next;
    }
    else
    {
        $barcode = $map{$temp};
    }

    $N++; 
    $seq="\@ST-E0:0:SIMULATE:8:0:0:$N"; 
    ## 1th
    print OUT "$seq 1:N:0:NAAGTGCT\n"; 
    ## 2th
    $S=<IN1>;
    print OUT "$barcode"."ATCGAGN"."$S";
    ## 3th
    $S=<IN1>;
    print OUT "$S"; 
    ## 4th
    $S=<IN1>;
    $S=~s/!/#/g;
    print OUT "FFFFFFFFFFFFFFFFFFFFFF#$S";  
    #          1234567890123456789012
    ## 1th
    $S=<IN2>; 
    print OUT3 "$seq 2:N:0:NAAGTGCT\n"; 
    ## 2th
    $S=<IN2>;
    print OUT3 "$S"; 
    ## 3th
    $S=<IN2>;
    print OUT3 "$S"; 
    ## 4th
    $S=<IN2>;
    $S=~s/!/#/g;
    print OUT3 "$S"; 

#    print OUT2 "$seq 1:N:0:NAAGTGCT\nTTCACGCG\n\+\nAAFFFKKK\n"; 
} 

close(IN1);
close(IN2);
close(OUT);
#close(OUT2);
