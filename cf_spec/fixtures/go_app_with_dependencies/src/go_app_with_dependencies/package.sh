cd ../..

export GOPATH=$PWD
export PATH=$GOPATH/bin:$PATH

go get github.com/tools/godep
godep get github.com/ZiCog/shiny-thing/foo

cd -
go install

mkdir -p Godeps/_workspace/src
cp -r ../github.com Godeps/_workspace/src/github.com