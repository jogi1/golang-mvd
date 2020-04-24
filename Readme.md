# golang-mvd
```
go get -u github.com/go-bindata/go-bindata/...
go get golang.org/x/tools/cmd/stringer
go generate
go-bindata data
go build
```

# output
output is handled via javascript run in a vm. if a file "runme.js" is in the same dir as the parser it will be used instead of the inbuild default (wich can be found in "data/default.js")
