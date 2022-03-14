#!/usr/bin/perl -w
use strict;

if ( @ARGV != 6 ) {
    print
"Example: perl split_barcode_PEXXX_42_reads_zyp.pl barcode.list barcode_RC.list read_1.fq.gz read_2.fq.gz 100 split_read \n";
    exit(0);
}

my $read_len = $ARGV[4];
my ( $n1, $n2, $n3, $n4, $n5 ) = ( 10, 6, 10, 6, 10 );
my %barcode_hash;
open IN, "$ARGV[0]" or die "cann't not open barcode.list";
my $n = 0;
while (<IN>) {

    #TAACAGCCAA 1
    #CTAAGAGTCC 2
    $n++;
    my @line       = split;
    my @barcode    = split( //, $line[0] );
    my $barcode_ID = $line[1];
    for ( my $num = 0 ; $num <= 9 ; $num++ ) {
        my @barcode_mis = @barcode;
        $barcode_mis[$num] = "A";
        my $barcode_mis = join( "", @barcode_mis );
        $barcode_hash{$barcode_mis} = $barcode_ID;
        @barcode_mis                = @barcode;
        $barcode_mis[$num]          = "G";
        my $barcode_mis = join( "", @barcode_mis );
        $barcode_hash{$barcode_mis} = $barcode_ID;
        @barcode_mis                = @barcode;
        $barcode_mis[$num]          = "C";
        my $barcode_mis = join( "", @barcode_mis );
        $barcode_hash{$barcode_mis} = $barcode_ID;
        @barcode_mis                = @barcode;
        $barcode_mis[$num]          = "T";
        my $barcode_mis = join( "", @barcode_mis );
        $barcode_hash{$barcode_mis} = $barcode_ID;
    }
}
close IN;
my $barcode_types = $n * $n * $n;

my $num1 = 0;
my $num2 = 0;
my ( $k1, $k2 );
my %hash;
my $split_barcode_num = 0;
open OUT1, "| gzip > $ARGV[5].1.fq.gz" or die "$!";
open OUT2, "| gzip > $ARGV[5].2.fq.gz" or die "$!";
open IN1,  "gzip -dc $ARGV[2] |"       or die "$!";
open IN2,  "gzip -dc $ARGV[3] |"       or die "$!";
my $LINE1;
my $LINE2;

while ( $LINE1 = <IN1>, $LINE2 = <IN2> ) {
    chomp($LINE1);
    chomp($LINE2);

#@V300019021L4C001R025000007/2
#CACTTGAACTATACTTTTGCCAACACTGTTAGCTACTAGGAGGTGACTGTTTCTCTGACATACTGCTATAAATGCCCATTCTTCATTTCAGGATCCTGTAGACTCGCTAGGGAGCGACGTTGTACGCCTTCCCACCGCTAGT
#+
#FFFFGFFFGFGFEF?FFFGGGGFFFFGGFFGGFGEGFGFFFFGFFFFGFFFFGGC=GEGFFFFGDGGGFFFFFFGGGFF@?FEFFFFG@GGFFFGGGF?FGFFGGFFGFF;(;&+,GGGFGFEGGG1.3,7*FFGFGGGGGF
    $num1++;
    if ( $num1 % 4 == 1 ) {
        $k1 = ( split /\//, $LINE1 )[0];
        $k2 = ( split /\//, $LINE2 )[0];
        if ( $k1 ne $k2 ) {
            print "error: $k1 not eq $k2 at the $num1 reads\n";
            exit;
        }
        next;
    }
    if ( $num1 % 4 == 2 ) {
        my $b1 = substr( $LINE2, $read_len,                         $n1 );
        my $b2 = substr( $LINE2, $read_len + $n1 + $n2,             $n3 );
        my $b3 = substr( $LINE2, $read_len + $n1 + $n2 + $n3 + $n4, $n5 );
        my $fq1= substr($LINE1,0,$read_len);
        my $fq2= substr($LINE2,0,$read_len);
        if ( exists $barcode_hash{$b1} && exists $barcode_hash{$b2} && exists $barcode_hash{$b3} ){
            my $id = $barcode_hash{$b1} . "_" . $barcode_hash{$b2} . "_" . $barcode_hash{$b3};
            $num2++;
            if ( !exists $hash{$id} ) {
                $split_barcode_num++;
                $hash{$id} = $split_barcode_num;
            }
            print OUT1 "$k1\#$id\/1\t$hash{$id}\t1\n$fq1\n";
            print OUT2 "$k2\#$id\/2\t$hash{$id}\t1\n$fq2\n";
        }else {
            print OUT1 "$k1\#0_0_0\/1\t0\t1\n$fq1\n";
            print OUT2 "$k2\#0_0_0\/2\t0\t1\n$fq2\n";
        }
        $LINE1 = <IN1>, $LINE2 = <IN2>;$num1++;
        print OUT1 "$LINE1";print OUT2 "$LINE2";
        $LINE1 = <IN1>, $LINE2 = <IN2>;$num1++;chomp ($LINE1);chomp ($LINE2);
        my $fq1_q=substr($LINE1,0,$read_len);
        my $fq2_q=substr($LINE2,0,$read_len);           
        print OUT1 "$fq1_q\n";print OUT2 "$fq1_q\n";
    }
}
close IN2;
close IN1;
close OUT1;
close OUT2;
$num1=$num1/4;
my $r = 100 *  $split_barcode_num/$barcode_types;
print "Barcode_types = $n * $n * $n = $barcode_types\n";
print "Real_Barcode_types = $split_barcode_num ($r %)\n";
print "Reads_pair_num  = $num1 \n";
$r = 100 *  $num2/$num1;
print "Reads_pair_num(after split) = $num2 ($r %)\n";
