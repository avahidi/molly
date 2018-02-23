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
   -disable <option>
   -enable <option>
   -outdir <output directory>   (default "output/extracted")
   -repdir <report directory>   (default "output/reports")
   -tagop <tagname:operation>   tag-op definition

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
