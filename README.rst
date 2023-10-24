Molly
=====

Molly (after Molly Hooper in Sherlock Holmes, not the drug) is an automated file analysis and extraction tool. It can search files for user-defined patterns and perform various actions when a match is found.

Molly was initially developed in the SECONDS (Secure Connected Devices) project for binary extraction from foreign firmware images.


Installation
------------

To build from source::

    sudo apt install golang build-essential git
    git clone https://github.com/avahidi/molly
    cd molly
    make && make test && make run

To build command-line tool from go::

    go install github.com/avahidi/molly/cmd/mh@latest


Example
-------

Lets run the Molly command-line utility "mh" on itself::

    $ mh -o output -p config.builtin=true ./mh
    SCAN RESULTS:
            * File mh (0 errors):
                    => ELF ELF_le ELF_x64

It seems mh recognizes itself being an ELF binary. Lets look at the generated report::

    $ ls output/
    match.json  mh  mh_molly.json  rules.json  summary.json

    $ cat output/summary.json
    ....
        "matches": { "mh": 1},
        "tags": {
                "elf": [ "mh" ],
                "executable": [ "mh"]
    ...


Rules
-----

Molly uses a rule database to store known patterns. The rules have a simple and familiar syntax, for example consider the ELF rule::

    rule ELF (tag = "executable,elf") {
        /* variables */
        var magic = String(0, 4);
        var class = Byte(4);
        var data = Byte(5);
        var version = Byte(6);
        /* conditions */
        if magic== "\x7FELF" && (version == 1)
            && (class == 1 || class == 2)
            && (data == 1 || data == 2);
    }


Rules can have children, which alllows multiple related file formats to be defined with minimal effort. For example::

    rule ELF_le (bigendian = false) : ELF {
        var machine = Short(18);
        if data == 1;
    }
    rule ELF_be (bigendian = true) : ELF {
        var machine = Short(18);
        if data == 2;
    }
    rule ELF_x86 (tag = "x86") : ELF_le {
        if machine == 0x0003;
    }

Rules can also define actions to be performed when a match is found. For example::

    rule ELF (tag = "executable,elf") {
        ...
        analyze("strings", "string_analysis");
        analyze("version", "");
    }

The resulting report file::

    $ ls output/
    match.json  mh  mh_molly.json  rules.json  summary.json

    $ cat output/mh_molly.json
    {
        "filename": "mh",
        ...
        "strings": [
            "runtime.throw",
            "compress/zlib.NewReader",
            "bufio.NewReader",
            "c=FrX",
        ...
        "possible-version": [
            "GLIBC_2.3.2",
            "GLIBC_2.2.5",
            "go1.18.1",
            ...

