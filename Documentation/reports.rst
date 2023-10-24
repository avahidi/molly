Reports
=======

Assume we ran the BSD ELF example on some FreeBSD binaries and got this output::

    $ mh -R bsdelf.rule -o output files/
    SCAN RESULTS:
        * File files/libavl.so.2 (0 errors):
             => bsdelf
        * File files/cat (0 errors):
             => bsdelf
        ...

This information will also be available in a machine-readable format in the output directory::

    $ cat  output/match.json
    {
        "matches": {
            "files/cat": [ "bsdelf" ],
            "files/libavl.so.2": [ "bsdelf"]
        }
        ...
    }

Furthermore, a detailed report is created for files with at least one match.
These reports will mirror the original file name and have a "_molly.json" suffix::

    cat output/files/libavl.so.2_molly.json
    {
        "depth": 0,
        "filename": "files/libavl.so.2",
        "information": {
        "checksum": "fa8b0b087b035e0937fe9358a93b20d502203b24def5b8316e717b0fdd648b43"
        },
        "matches": [
            {
                "Name": "bsdelf",
                "Vars": {
                    "magic": "ELF",
                    "osabi": 9,
                }
            }
        ],
        ...
    }


Molly also keeps track of file hierarchies.
For example if the above file was initially stored in files/container.zip, the following would appear instead::

    $ cat output/files/container.zip_/libavl.so.2_molly.json
    {
        "depth": 1,
        "parent": "files/container.zip",
        "filename": "output/files/container.zip_/libavl.so.2",
        ...
    }



