Molly
=====

Molly (after Molly Hooper in Sherlock Holmes, not the drug) is an automated file analysis and extraction tool. It can search files for user-defined patterns and perform various actions when a match is found.  

Molly comes with a number of operators for analyzing and files in addition to a simple API for adding custom ones.

Molly was initially developed in the SECONDS (Secure Connected Devices) project for binary extraction from foreign firmware images.

Installation
------------

Binaries are found on the `download page <https://github.com/avahidi/molly/downloads/>`_, but might be slightly out of date.

To build from source::

    sudo apt install golang build-essential git
    git clone https://github.com/avahidi/molly
    cd molly
    make && make test && make run

Rules
-----

Molly uses a rule database to store known patterns. The rules have a simple and familiar syntax, for example the following will recognize ZIP files)::

    rule ZIP (bigendian = false, tag = "archive") {
        var header = String(0, 4); /* extract 4-byte string at position 0 */
        var csize = Long(18);      /* extract 32-bit at position 18 */
        var usize = Long(22);
        if header == { 'P', 'K', 0x05, 0x06} || header == {'P', 'K', 0x03, 0x04};
        extract("zip", "");       /* apply the ZIP extractor on this file */
    }

For more detailed information refer to the `manual <manual.pdf>`_.
