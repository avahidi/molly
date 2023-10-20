
Writing molly rules
===================


.. image:: elf.png

Consider the ELF file format shown here. Given this information we could define rules to, for example, identify FreeBSD binaries:

- The file must start with the correct four "magic" bytes,
- The byte at offset 7 (OS ABI) must contain 9.

The syntax for defining this format in Molly is::

    // elfbsd.rule
    rule bsdelf {
        // variables
        var magic = String(0, 4);
        var osabi = Byte(7);

        // conditions
        if magic == {0x7f, 0x45, 0x4c, 0x46};
        if osabi == 0x09;
    }



Environment variables
---------------------

Environment variables hold additional data from the environment including the target file.
In a rule, such variables have a "$" prefix::

    rule biggofile {
        if $filesize > 4096;
        if $ext == ".go";
        printf("%s is one big Go file...\n", $filename);
    }

To avoid confusion with shell environment variables curly brackets are instead used on the command-line::

    $ molly -r 'rule cfiles { if $ext == ".c"; }' -on-rule "cfiles:gcc {filename} -o {newfile:compiled.o}" src/


The following environment variables are available with rules:

========================  =======  ========================================
Variable                  Type     Description
========================  =======  ========================================
filename                  string   Name of the currently processed file, e.g. */dir/file.c"*
shortname                 string   Short file name, e.g. *"file.c"*
basename                  string   Base file name, e.g. *"file"*
ext                       string   File extension, e.g. *".c"*
dirname                   string   File directory, e.g. *"/dir"*
filesize                  int64    File size
parent                    string   Name of parent file (or "" if root)
depth                     int      Depth of file in extraction tree, 0 if root
num_matches               int      number of matches for this file so far
num_errors                int      number of errors encountered for this file so far
num_logs                  int      number of logs generated for this file so far
newfile[:suggestedname]   string   produce new file (see note)
newdir[:suggestedname]    string   produce new directory (see note)
========================  =======  ========================================


Note that newfile/newdir are only available on the command-line only. Furthermore, the files
and folder created with these are fed back into Molly for analysis.

Metadata
--------

Additional configuration parameters in rules are presented as metadata:

=========  ==========  ==============  ==================  =========================
Metadata   Type        Default value   Usage               Description
=========  ==========  ==============  ==================  =========================
tag        string                      rules               associate rule with a tag
bigendian  boolean     true            rules / operators   specify endianness
pass       0/1/2       0               rules               scanner pass
=========  ==========  ==============  ==================  =========================


Metadata must always be a static value. In the example below, *bigendian* and *tag* are metadata::

    rule example (tag ="test", bigendian = false ) {
        var a = Long(0); // little endian, by rule metadata
        var b = Long(0, bigendian = true); // bigendian, overides rule
    }


The rules are checked in one of three passes (pass 0 is the default).
This allows rules to take action based on previous matches.
For example the following rule will invoke binwalk as a last resort if no matches have been found by molly so far::

    rule unknown (pass = 1) {
        // NOTE: condition to detect endless-loop in binwalk omitted for brevity!
        if $num_matches == 0;   // nothing found by previous rules
        var d = dir("from_binwalk");  // store results here
        system("binwalk -C %s -d 4 -e -M %s", d, $filename);
    }




Hierarchical rules
------------------

Defining related rules as a hierarchy can lead to shorter and simpler rule definitions while minimizing chance of mistakes and at the same time improving performance.
In the following example multiple ELF architectures are covered by use of rule hierarchies::

    rule ELF {
        var magic = String(0, 4);
        var class = Byte(4);
        var data = Byte(5);
        var version = Byte(6);
        var osabi = Byte(7);
        if magic== "\x7FELF" && (version == 1) &&
            (class == 1 || class == 2) && (data == 1 || data == 2);
    }
    rule ELF_be (bigendian = true) : ELF {
        if data == 2;
    }
    rule ELF_le (bigendian = false) : ELF {
        var machine = Short(18);
        if data == 1;
    }
    rule ELF_x86 : ELF_le {
        if machine == 0x0003;
    }
    rule ELF_arm64 : ELF_le {
        if machine == 0x00B7;
    }


Metadata is also inherited, hence in this example both ELF\_x86 and ELF\_arm64 are little-endians.


Operators
=========

Operators are functions that can be called within rules.
The most common operators are the primitive operators for reading data.
Other operators mainly operate on existing variables or the whole file.

===================================================  ======================================================
Operator                                             Description
===================================================  ======================================================
*Primitive operations*
-----------------------------------------------------------------------------------------------------------
uint8 **Byte** (offset int)                          read 1 byte from offset
uint16 **Short** (offset int)                        read 2 bytes
uint32 **Long** (offset int)                         read 4 bytes
uint64 **Quad** (offset int)                         read 8 bytes
string **String** (offset, size int)                 read a byte octet from given offset
string **StringZ** (offset, maxsize int)             read a zero-terminated string with given max size
*String operations*
-----------------------------------------------------------------------------------------------------------
bool **stricmp** (string, string)                    string compare, ignore case
bool **strstr** (string, string)                     find string in text
bool **strcasestr** (string, string)                 same as strstr but case-insensitive
bool **strsuffix** (string, string)                  check if text ends with some string
bool **strprefix** (string, string)                  check if text starts with some string
int **strlen** (string)                              string length
int64 **strtol** (string)                            convert string to number
string **strupper** (string)                         upper-case string
string **strlower** (string)                         lower-case string
*Formatting*
-----------------------------------------------------------------------------------------------------------
string **printf** (string, ...any)                   Standard printf to stdout (Go syntax)
string **sprintf** (string, ...any)                  Standard printf to string (Go syntax)
*New Input*
-----------------------------------------------------------------------------------------------------------
string **dir** (string)                              Create a new directory
string **file** (string)                             Create a new file
*Miscellaneous*
-----------------------------------------------------------------------------------------------------------
[]uint8 **checksum** (type string, ...uint64)        Checksum file or slice
int **len** (any)                                    Return length of item
string **epoch2time** (int64)                        Convert UNIX epoch to a date string
bool **has** (type string, string)                   Target has the following data or metadata
*Actions*
-----------------------------------------------------------------------------------------------------------
string **system** (command string, ...any)           Execute shell commands
string **analyze** (format, file string, ...any)     Perform analysis on file
string **extract** (format, file string, ...uint64)  Extract file or file slice
===================================================  ======================================================



Actions and other complex operators
===================================

Shell commands
--------------

The *system* operator executes a shell command. It uses the same formatting syntax as printf/sprintf::


    rule squashfs {
        ...
        var dir = dir("unpacked_stuff");
        system("unsquashfs -n -no -f -d %s %s", dir, $filename);
    }


Executing shell commands is regarded a dangerous operation and should be avoided. A better solution is to enhance Molly with your own operators, analyzers and extractors.
To protect the user from harm, the system action is disabled and must be enabled using the parameter "-p perm.execute=true".



Checksums
---------
The *checksum* operator computes a checksum or hash over a range of bytes. For example::

    rule example {
        ...
        checksum("sha256", 0, 1024); /* SHA-256 for the first KB */
    }

It currently supports the following types: sha256, sha128, sha1, md5, crc32, crc32-ieee, crc32-castagnoli, crc32-koopman, crc64, crc64-iso, crc64-ecma.

Analyzers
---------
The *analyze* operator performs some type of analysis on the current file. For example::

    rule DalvikDex (bigendian = false) {
        ...
        analyze("dex", "my-dex-analysis");
    }

The outcome of this analysis will be found in the generated report.

Currently the following analyzers are supported:

* strings: String extraction
* version: Version extraction from strings
* histogram: Generate byte histogram
* elf: ELF analyzer
* dex: Android DEX analyzer


Extractors
----------

The *extract* operator extracts data from the target file. For example::

    rule jffs2 (tag = "filesystem", bigendian = false) {
        ...
        extract("jffs2", "jffs2");
    }

The currently supported formats are binary, tar, MBR, cramfs, JFFS2, zip, gz, CPIO and uImage.

The binary extractor can also operate on file *slices*.
In this context a file *slice* is a subset of a file and is defined by the pair *(offset, length)*.

For example, the following will extract bytes 10, 11, 20 and 21::

        extract("binary", "mydata", 10, 2, 20, 2);
