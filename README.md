[![Go Lang](http://kavehmz.github.io/static/gopher/gopher-front.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/kavehmz/callp/callp?status.svg)](https://godoc.org/github.com/kavehmz/callp/callp)
[![Build Status](https://travis-ci.org/kavehmz/callp.svg?branch=master)](https://travis-ci.org/kavehmz/callp)
[![Coverage Status](https://coveralls.io/repos/github/kavehmz/callp/badge.svg?branch=master)](https://coveralls.io/github/kavehmz/callp?branch=master)

# callp
Idea behind callp is to create a daemon which will call a Perl script.
Perl script handles the legacy code which does financial market pricing.

callp start several perl-pricer concurrently and control them.

This is a temporary solution for a code which its refactoring is not an option but needs some sort of concurrency support.

callp uses Redis atomic nature and calls like SETNX to communicate with another old Perl Websockets service to get new pricing tasks.
