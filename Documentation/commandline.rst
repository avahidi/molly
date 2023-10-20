

Command-line options
====================

The command-line format is::

    molly [OPTIONS] files

Where *files* are zero or more files (or directories) to be analyzed. Options are


=====================  ==========================================
Option                 Description
=====================  ==========================================
-R rule file           Rule file (or directory)
-r rule text           In-line rule
-o dir                 Output directory (default "output")
-p param=value         Set Molly parameter
-on-rule rule:command  Command to run when a rule match is found
-on-tag tag:command    Command to run when a tag match is found
-version               Show version number and exit
-h                     Help information
-H                     Extended help information
=====================  ==========================================


Available parameters are:

=====================  ==============  ===========
Name                   Default value   Description
=====================  ==============  ===========
config.maxdepth        12              Max scan depth
config.standardrules   true            Load standard rule library
perm.create-file       true            Allow Molly to create new files
perm.execute           false           Allow Molly to execute external tools
option.verbose         false           Be verbose
=====================  ==============  ===========

Molly comes with a small set of standard rules which can be excluded by setting *config.standardrules* to *false*.


The *-on-rule* and *-on-tag* parameters allows execution of external commands when a rule or tag match is seen::

    $ touch file1
    $ molly -r "rule any{ }" -on-rule "any:ls -l {filename}" file1
    RULE any on file1: -rw-rw-r-- 1 mh mh 6 mar  0 13:55 file1
    $ echo hello > file1
    $ molly -r "rule any (tag = \"text\") { }" -on-tag "text: cat {filename}" file1
    TAG text on file1: hello


Note that *{filename}* is a molly environment variable.

