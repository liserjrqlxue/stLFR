use strict;
use File::Basename qw'basename';
my $list   = shift;
my $outdir = shift;
open LIST, $list;
while (<LIST>) {
  chomp;
  my $fq1    = $_;
  my $fq2    = s/_R1_/_R2_/r;
  my $outfq1 = basename($fq1);
  my $outfq2 = basename($fq2);
  print
    "/share/udata/wangyaoshen/src/HPC_chip/tools/SOAPnuke filter -1 $fq1 -2 $fq2 -o $outdir -C $outfq1 -D $outfq2 -G\n";
}
