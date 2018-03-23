molly
=====

molly is an automated file analysis and extraction tool.


Molly was initially developed in the `Seconds <http://www.secondssolutions.com/>`_
project for binary extraction from firmware images.


Installation
------------

molly is written in `Go <https://golang.org>`_ and has no external dependencies.
To install Go on Ubuntu 16.04 LTS::

   sudo apt install golang build-essential
   export GOPATH=$HOME/go
   export PATH=$PATH:$GOPATH/bin
   mkdir $GOPATH

Now download and build Molly::

    go get -u bitbucket.org/vahidi/molly/...
    go install bitbucket.org/vahidi/molly/...

For development builds we use make::

    cd $GOPATH/src/bitbucket.org/vahidi/molly
    make
    make test
    make run
    ...


Download
--------

If you prefer to download pre-built binaries, go to the
`download page <https://bitbucket.org/vahidi/molly/downloads/>`_ .


Usage
-----

The command-line format is::

    molly [options] <input files>

Options are::

   -h	                        help information
   -V	                        show version number
   -v	                        be verbose
   -R <rule files>              rules to load
   -r <inline rule>             inline rule string
   -disable <option>
   -enable <option>
   -outdir <output directory>   (default "output/extracted")
   -repdir <report directory>   (default "output/reports")
   -on-rule <rulename:cmd>      rule match operation definition
   -on-tag <tagname:cmd>        tag match operation definition

A small set of default rules are provided in the distribution.


Rule format
-----------

Rules have the following format::

   rule <rule name> [(<metadata>, ...)] [ : <parent name>] {
       <variables>
       <conditions>
       <actions>
    }

For example::

    rule ZIP (bigendian = false, tag = "archive") {
        var header = String(0, 4); /* extract 4-byte string at position 0 */
        var csize = Long(18);      /* extract 32-bit at position 18 */
        var usize = Long(22);
        if header == { 'P', 'K', 0x05, 0x06} || header == {'P', 'K', 0x03, 0x04};
        extract("zip", "");       /* apply  the ZIP extractor on this file */
    }


Match actions
-------------

You can define additional molly actions using the "-on-tag" and "-on-rule"::

    $ echo hello > file1
    $ molly -r "rule any{ }" -on-rule "any:ls -l {filename}" file1
    ...
    -rw-rw-r-- 1 mh mh 6 mar  6 13:55 file1
    $ molly -r "rule any (tag = \"text\") { }" -on-tag "text: cat {filename}" file1
    ...
    hello

Note that you can "{newXXX[:name]}" to track files generated externally::

    $ molly -r "rule cfiles { ...  }" -on-rule "cfiles:gcc {filename} -o {newfile:.o}" file1.c




API
---

UNDER CONSTRUCTION :)


FAQ
---


Why the name?
~~~~~~~~~~~~~

molly was named after Molly Hooper, from the BBC TV-series Sherlock.
According to Wikipedia "Molly Hooper [...] is a 31-year-old specialist registrar
working in the morgue at St Bartholomew's Hospital [...]". This seemed appropriate
for a software used to dissect long dead binaries.
