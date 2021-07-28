go build main.go
rm -r -f pack-output
mkdir pack-output
cp ./main ./pack-output/youfile
cp -a ./pack/. ./pack-output