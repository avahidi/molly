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

Actions and operators
---------------------

The following extraction operators can be used within rules::

    // extract functions
    String(offset, size int)
	StringZ(offset, maxsize int)
	Byte(int offset)
	Short(int offset)
	Long(int offset)
	Quad(int offset)

Molly also provides a large number if built-in operators and actions
ranging from simple string manipulation functions (e.g. strlen) to complex
analysis and extraction actions.

Users can also register their own actions grammatically. Furthermore some complex actions
can register additional formats or algorithms. See the API section for more information.


Special variables
-----------------

The following special variables can be accessed in rules and match-actions (see below):
filename, shortname, basename, dirname, ext (extension), filesize, parent and depth.

Within rules special variables have the "$" prefix, for example::

    rule biggofile {
        if $filesize > 4096;
        if $ext == ".go";
        printf("%s is one big Go file...\n", $filename);
    }


Match actions
-------------
In addition to actions defined in rules one can also define match actions
using the "-on-tag" and "-on-rule" command line parameters::

    $ echo hello > file1
    $ molly -r "rule any{ }" -on-rule "any:ls -l {filename}" file1
    ...
    -rw-rw-r-- 1 mh mh 6 mar  6 13:55 file1
    $
    $ molly -r "rule any (tag = \"text\") { }" -on-tag "text: cat {filename}" file1
    ...
    hello

Note that special variables use the "{variable}" format to avoid confusion
with shell variables. In addition, match actions can access two new variables
"{newfile[:suggestedname]}" and "{newdir[:suggestedname]}" for cases where
the action will produce new files that one wants to feed back to molly for analysis::

    $ molly -r 'rule cfiles { if $ext == ".c"; } -on-rule "cfiles:gcc {filename} -o {newfile:compiled.o}" src/


Order of execution
------------------

Conditions and actions are executed in the order they appear while variables
are evaluated when needed. This means you can optimize rules by placing
simpler conditions first.

Furthermore, if an action fails the subsequent actions will not be executed.
There are two exceptions to this: if the action is preceded by a '-' or a '+'
errors are ignored. In the latter case molly will also stop executing subsequent
actions if this action succeeds. Example::

    rule unknown {
        -printf("I don't know what %s is", $filename);  // this can fail
        +extract("zip", ""); // could be a zip?         // if this doesnt fail...
        extract("tar", ""); // or maybe a tar?          // ... this will run
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
