# gut

`cut` clone in Go

```sh
./gut -d "," -f 1,2,3-5 -s -p 10 in.csv
```

## Install

```sh
go build
```

## Test file

```sh
url=ftp://orengoftp.biochem.ucl.ac.uk/cath/releases/all-releases/v4_3_0/cath-classification-data/cath-domain-list-v4_3_0.txt
curl -o test.txt $url
sed '/^\#/d' test.txt | tr -s " " | tr " " "\t" > test.tsv
du -sh test.tsv # 20M
wc -l test.tsv # 500238
```

## Correctness of concurrency in parallel algorithm

The order of lines in the output using the parallel algorithm is nondeterministic, so we must sort the lines first.

```sh
./gut -f 1,3,5,7      test.tsv | sort | shasum -a 256
./gut -f 1,3,5,7 -p 6 test.tsv | sort | shasum -a 256
cut   -f 1,3,5,7      test.tsv | sort | shasum -a 256
```

## Benchmark against `cut`

```sh
time ./gut -f 1,3,5,7      test.tsv > /dev/null
time ./gut -f 1,3,5,7 -p 6 test.tsv > /dev/null
time cut   -f 1,3,5,7      test.tsv > /dev/null
```

Results:

```console
❯ time ./gut -f 1,3,5,7      test.tsv > /dev/null
time ./gut -f 1,3,5,7 -p 6 test.tsv > /dev/null
time cut   -f 1,3,5,7      test.tsv > /dev/null
./gut -f 1,3,5,7 test.tsv > /dev/null  0.39s user 0.04s system 106% cpu 0.407 total
./gut -f 1,3,5,7 -p 6 test.tsv > /dev/null  2.11s user 0.52s system 368% cpu 0.713 total
cut -f 1,3,5,7 test.tsv > /dev/null  0.43s user 0.01s system 99% cpu 0.450 total
```