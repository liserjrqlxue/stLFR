# stLFR分析流程

## 分析步骤

### 1. 下机数据预处理

1. 命令：

```
mkdir -p $workdir/fq/$id
stLFR/util/split10x/split10x -fq1 $fq1 -fq2 $fq2 -prefix $workdir/fq/$id/$id
```

### 2. `longrager` 分析

1. 命令：

```
export PATH=/path/to/longranger:$PATH
ref=/path/to/refdata-hg19
gatk=/path/to/gatk.jar
cd $workdir
longranger wgs --id=$id --fastqs=$workdir/fq/$id --reference=$ref --vcmode=gatk:$gatk --localcores=96 --localmem=378
```
