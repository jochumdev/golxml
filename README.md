## A fast replacement for encoding/xml.Unmarshal() based on Libxml2 for the Go programming language

### Requirements

* [Go](http://golang.org/doc/install) >=1.0.2
* [gokogiri](https://github.com/moovweb/gokogiri) the libxml wrapper
* [gocheck](http://labix.org/gocheck) for tests/benchmarks 


### Usage

See the [tests](https://github.com/pcdummy/golxml/blob/master/xml/xml_test.go) for examples 


### Installation

Install with **go get** (make sure **$GOPATH** is not set to install in **$GOROOT**)

	$ go get -u github.com/pcdummy/golxml/xml (-u flag for "update")


Run the tests

	$ cd $GOROOT/src/pkg/github.com/pcdummy/golxml/xml
	$ go test *.go
	$ go test -gocheck.b *.go

### Known issues

* No namespace support
* No "omitempty" support
* Gokogiri does not compile with gccgo

### TODO list

	* Add "omitempty" support
	* Add support for gccgo