molly
=====

molly is a firmware identification and extraction tool. It can be used to analyze
unknown binaries or automate extraction of data from various data formats.


Molly was developed in the `Seconds <http://www.secondssolutions.com/>`_
project for binary extraction from firmware images.



Installation
------------

molly is written in `Go <https://golang.org>`_ and has no external dependencies.
To install Go on Ubuntu 16.04 LTS::

   sudo apt install make golang
   export GOPATH=$HOME/go
   export PATH=PATH:$GOPATH/bin
   mkdir $GOPATH $GOPATH

Now download and build Molly::

    go install bitbucket.org/vahidi/molly/...


Usage
-----




misc
---

Why the name?
~~~~~~~~~~~~~

molly was named after Molly Hooper, from the BBC TV-series Sherlock.

According to Wikipedia "Molly Hooper [...] is a 31-year-old specialist registrar
working in the morgue at St Bartholomew's Hospital [...]". This seemed appropriate
for a software used to dissect long dead firmware images.
