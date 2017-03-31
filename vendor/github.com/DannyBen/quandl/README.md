Go Quandl
==================================================

[![Build Status](https://travis-ci.org/DannyBen/quandl.svg?branch=master)](https://travis-ci.org/DannyBen/quandl)
[![GoDoc](https://godoc.org/github.com/DannyBen/quandl?status.png)](http://godoc.org/github.com/DannyBen/quandl)

---

This library provides easy access to the 
[Quandl API](https://www.quandl.com/help/api) 
using the Go programming language.

The full documentation is at:  
[godoc.org/github.com/DannyBen/quandl](http://godoc.org/github.com/DannyBen/quandl)

---

Install
--------------------------------------------------

	$ go get github.com/DannyBen/quandl


Features
--------------------------------------------------

* Supports 3 call types to Quandl: `GetSymbol`, `GetList` and `GetSearch`.
* Returns either a native [Go object](https://github.com/DannyBen/quandl/blob/master/quandlResponseTypes.go), or a raw (CSV/JSON/XML)
  response.
* Built in cache handling.


Usage
--------------------------------------------------

Basic usage looks like this:


```go
quandl.APIKey = "YOUR KEY"
data, err := quandl.GetSymbol("WIKI/AAPL", nil)
```

and will return a native Go object. To use the data in the
response, iterate through its Data property:

```go
for i, item := range data.Data {
  fmt.Println(i, item[0], item[2])
}
```

To receive a raw response from Quandl (CSV, JSON, XML)
you can use:

```go
data, err := quandl.GetSymbolRaw("WIKI/AAPL", "csv", nil)
```

To pass options to the Quandl API, use something like this:

```go
v := quandl.Options{}
v.Set("trim_start", "2014-01-01")
v.Set("trim_end", "2014-02-02")
data, err := quandl.GetSymbol("WIKI/AAPL", v)
```

More examples are in the 
[quandl_test file](https://github.com/DannyBen/quandl/blob/master/quandl_test.go)
or in the 
[official godoc documentation](http://godoc.org/github.com/DannyBen/quandl#pkg-examples)


Development
--------------------------------------------------

Before running tests, set your API key in an environment variable.

	$ export QUANDL_KEY=your_key_here
	$ go test -v

