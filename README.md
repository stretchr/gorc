gort - "go test ./...", and then some
===================================

gort ( **Go** **R**ecursive **T**esting ) is a go program that installs test dependencies and/or runs tests, recursively, starting from the working directory.

gort allows you to run all the tests in a multi-package directory structure with one command. As it works, it will display a '.' for each successful test and print out any failures. At the end of the run, gort will give you a summary of tests run, tests succeeded and tests failed.

gort pairs wonderfully with our testing framework, [Testify](http://github.com/stretchrcom/testify).

For an introduction to writing test code in Go, see the [Go Testing Documentation](http://golang.org/doc/code.html#Testing).

Installation
============

Before installation, please ensure that your GOPATH is [set properly](http://golang.org/doc/code.html#tmp_2).

To install gort, use `go get`:

    go get github.com/stretchrcom/gort
	
`go get` should install gort to $GOPATH/bin. In some cases, go will not use $GOPATH, and instead attempts to install to $GOBIN. If this happens, you can grant it permission to do so, or simply build and copy gort to $GOPATH/bin manually.


Usage
=====

Using gort is easy. Just execute:

	gort
or
	
	gort test

gort will recurse the directory structure and run `go test -i` & `go test` for each directory that contains tests.

If there is a directory that contains tests you don't wish to run, simply exclude it:

	gort exclude testify
	
Now, when you run tests, the directory "testify" will not be recursed, and no tests inside it or its subdirectories will be run.

gort has some more commands that are not listed here. To see them all, run:

	gort help


------

Contributing
============

Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, we ask that you please include steps to reproduce the issue so we can see it on our end also!


Licence
=======
Copyright (c) 2012 Mat Ryer and Tyler Bunnell

Please consider promoting this project if you find it useful.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.