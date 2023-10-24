
Performing actions
==================


Actions from command-line
-------------------------

You may want to automatically perform an action once a match has been found.
From the command line this is done using "-on-rule" parameter as demonstrated below::

    $ mh -R bsdelf.rule -o output \
        -on-rule "bsdelf:echo Found {filename} with {filesize} bytes" files
    ...
    RULE bsdelf on files/libavl.so.2: Found files/libavl.so.2 with 14896 bytes
    RULE bsdelf on files/cat: Found files/cat with 23648 bytes
    ...


The "-on-tag" parameter works similarly, but is triggered by a tag associated by one or more rules.


Actions from rules
------------------

It is also possible to use operators that perform an action from within the rule.
This method has the benefit of having access to internal Molly data such as rule variables.

For example, assume we want to look at ELF headers for each detected ELF file.
This is done with the *system* operator.

Note: since *system* is a potentially dangerous operator, it must be enabled with the command.line option "-p perm.execute=true".

::

    rule bsdelf {
        ...
        system("objdump -h %s", $filename);
    }


Note that this action will execute silently and not generate any output. Fortunately this is easy to fix::

    rule bsdelf {
        ...
        var output = system("objdump -h %s", $filename);
        printf("ELF HEADERS:\n%s\n", output);
    }

Molly provides a rich set of action operators including complex data analysers and extractors.

