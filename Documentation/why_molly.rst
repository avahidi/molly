
Background and motivation
=========================

*Molly was initially developed to analyse legacy firmware images in the SECONDS project, with the following justification:*

Connected embedded systems are becoming more common, which has unfortunately resulted in a new class of possibly insecure devices.
Given the huge numbers and diversity of such devices, security researchers have turned to automated tools for vulnerability analysis.
Some notable attempts in large-scale automated vulnerability analysis are [Feng:2016:SGB:2976749.2978370]_ , [Pewny:2015:CBS:2867539.2867681]_ , [discovre]_ , [f8e193bc4a584ee890cf350f242d8621]_

A first obstacle such tools may encounter is extracting software components from the device, or more commonly from firmware images published by the vendors.
Firmware images for embedded systems are often container formats that include components such as
kernels, bootloaders and various types of filesystems. These filesystems are themselves container formats for binaries and data files.
Sometimes the container formats are proprietary or undocumented variation of standard formats.
Hence an important role of a file analysis tools may be to understand and extract data from such container formats.

There exist a number of tools for analysis and extraction of firmware. Molly is one such tool.


Existing tools
--------------
The *file* utility from UNIX operating systems [Ritchie74theunix]_ is one of the earliest file analysis software still being used today. It attempts to identify a file using three tests (filesystem, magic database and language). It is able to identify simple classes of files such as images and documents::

    $ file main.tex Makefile biblio.bib main.pdf
    main.tex:   LaTeX 2e document, ASCII text, with very long lines
    Makefile:   makefile script, ASCII text
    biblio.bib: ASCII text, with very long lines
    main.pdf:   PDF document, version 1.5



*Binwalk* is a more recent tool that uses an extension of the magic format database used by *file*.
It is able to identify many binaries and container formats (e.g. file systems and firmware images) using a fairly large magic database [binwalk]_. An example of a magic file is seen below::

    # Sample magic pattern for a Microsoft executable.
    # Note the "(0x3c.l)" which is a 32-bit pointer reference.
    0         string  MZ\0\0\0\0\0\0  Microsoft executable,
    >0x3c     lelong  <4              {invalid}
    >(0x3c.l) string  !PE\0\0         MS-DOS
    >(0x3c.l) string  PE\0\0          portable (PE)



*YARA* is another tool commonly used for large scale binary analysis [yara]_. It uses a rule database containing textual or binary patterns to identify classes of files. As seen below, YARA rules are human-readable while magic rules are mainly formatted for machine interpretation::

    rule maybeexecutable {
        strings:
            $mz="MZ"
        condition:
            ($mz at 0)
    }


These tools also provide a number of options for integration in a automated system and various methods to extend their core functionality. For example modern versions of *file* provide the *libmagic* library
[libmagic]_ and  [binwalk]_ and *YARA* both provide APIs for adding custom functionality.

It should be mentioned that a virus scanner is normally not seen as a generic binary analysis tool but a highly specialized data analysis software. However, some virus scanners such as the Clam anti virus build upon a general purpose file analysis engine
[clamav]_.

Why a new tool?
---------------

Molly improves on existing tools in a number of areas:

#. Molly has a simple human-readable rule format that resembles C / Java and makes definition of chains of pointers and conditions easier and more efficient,
#. Molly can perform complex actions when a match is found,
#. Molly has a powerful plugin system,
#. Molly can generate machine-readable output (JSON).

At the same time, it does have a number of disadvantages:

#. Molly currently does not have a comprehensive rule database,
#. Molly is not the best tool for analysing unknown binaries.

As such, we recommend using binwalk and Molly together, the former for the initial analysis and the latter for large-scale identification and extraction.


.. [Ritchie74theunix] The UNIX Time-Sharing System, D. M. Ritchie and K. Thompson
.. [YARA] https://github.com/plusvic/yara
.. [Feng:2016:SGB:2976749.2978370] Scalable Graph-based Bug Search for Firmware Images, Feng, Qian and Zhou, Rundong and Xu, Chengcheng and Cheng, Yao and Testa, Brian and Yin, Heng
.. [Pewny:2015:CBS:2867539.2867681] Cross-Architecture Bug Search in Binary Executables, Pewny, Jannik and Garmany, Behrad and Gawlik, Robert and Rossow, Christian and Holz, Thorsten
.. [discovre] discovRE: Efficient Cross-Architecture Identification of Bugs in Binary Code, Eschweiler, Sebastian and Yakdan, Khaled and Gerhards-Padilla, Elmar
.. [f8e193bc4a584ee890cf350f242d8621] Compiler-Agnostic Function Detection in Binaries, D.A. Andriesse and J.M. Slowinska and H.J. Bos
.. [binwalk] https://github.com/ReFirmLabs/binwalk
.. [libmagic] http://mx.gw.com/pipermail/file/2003/000043.html"
.. [clamav] https://www.clamav.net
