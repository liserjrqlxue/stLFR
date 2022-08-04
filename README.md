# stLFR分析流程

## 下机数据预处理

1. 命令：

```
mkdir -p fq/$id
/jdfstj1/B2C_RD_P2/USR/wangyaoshen/src/github.com/liserjrqlxue/stLFR/util/split10x/split10x -fq1 $fq1 -fq2 $fq2 -prefix fq/$id/$id
```

2. 参考脚本：

```
/ifstj1/B2C_RD_S1/USER/fangzhonghai/project/stLFR/20220702/prepare.sh
```

## longrager分析

1. 命令：

```
longranger wgs --id=$id --fastqs=fq/$id --reference=$ref --vcmode=gatk:/path/to/GenomeAnalysisTK.jar --localcores=96 --localmem=378
```

2. 参考脚本：

```
/ifstj1/B2C_RD_S1/USER/fangzhonghai/project/stLFR/20220702/target_wgs.exit.sh
```
