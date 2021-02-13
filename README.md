# gut
Go clone of cut

```sh
./gut -w -f 1,2,3-5 -p 10 in.file
```
## Test file

```sh
url=ftp://orengoftp.biochem.ucl.ac.uk/cath/releases/all-releases/v4_3_0/cath-classification-data/cath-domain-list-v4_3_0.txt

curl -o test.txt $url

sed -i .bak '/^\#/d' test.txt

cat test.txt | tr -s " " | tr " " "\t" > test.tsv
```

## Test correctness of concurrency in parallel algorithm

```sh
./gut -w -f 1,3,5 test.tsv | sort | shasum -a 256
./gut -w -f 1,3,5 -p 10 test.tsv | sort | shasum -a 256
```

## Benchmark against `cut`

```sh
cut -f 1,3,5,7 test.tsv | shasum -a 256
./gut -f 1,3,5,7 test.tsv | shasum -a 256

time cut -f 1,3,5,7 test.tsv > /dev/null
time ./gut -f 1,3,5,7 test.tsv > /dev/null
```