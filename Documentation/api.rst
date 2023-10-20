
Using the library
=================

Molly can be used as a library. A simple example on how this is used is shown below::

    package main
    import (
        "log"
        "bitbucket.org/vahidi/molly"
    )
    func main() {
        m := molly.New()
        m.Config.OutDir = "/tmp/Molly"
        if err := molly.LoadRules(m, "myrule.rule"); err != nil {
            log.Fatal(err)
        }
        if err := molly.ScanFiles(m, "unknownfile.bin"); err != nil {
            log.Fatal(err)
        }
        report := molly.ExtractReport(m)
        for _, f := range report.Files {
            log.Printf("%s, %d bytes. %d errors, %d matches\n",
                f.Filename, f.Filesize,  len(f.Errors), len(f.Matches))
        }
    }


An important reason for using Molly in this fashion is the ability to add custom functionality.


Custom operators
----------------

To simplest way to extend Molly is to add a custom operator. These can then be called from rules to perform complicated tasks.
For example, the following code implements AES code block decryption::

    import (
        "fmt"
        "crypto/aes"
        "bitbucket.org/vahidi/molly/types"
        "bitbucket.org/vahidi/molly/operators"
        "bitbucket.org/vahidi/molly"
    )
    // Signature must match func(e *types.Env, args ...interface{}) (interface{}, error)
    // Molly will attempt to convert between similar types but this may not always work
    func decrypt(e *types.Env, key, data []byte) ([]byte, error) {
        block, err := aes.NewCipher(key)
        if err != nil {
            return nil, err
        }
        bs := block.BlockSize()
        if len(data) % bs != 0 {
            return nil, fmt.Errorf("AES data must be padded to block size %d", bs)
        }
        dst := make([]byte, len(data))
        for i := 0; i < len(data); i += bs {
            block.Decrypt(dst[i:i+bs], data[i:i+bs])
        }
        return dst, nil
    }
    func main() {
        operators.Register("decrypt", decrypt);    // must happen before loading rules
        m := molly.New()
        ...
    }


After registration the new operator can be used in rules::

    rule example {
        var key = String(0, 32);                   // 256 bits keys for AES-256
        var ciphertext = String(32, 32);           // two blocks of data
        var cleartext = decrypt(key, ciphertext);
        ...
    }


Custom analyzers
----------------

Molly provides a number of default analyzers such as *strings* and *histrogram*. Additional analyzers can be added using the API::

    // Signature should be func(string, io.ReadSeeker, ...interface{}) (interface{}, error)
    // By convention, analyzers return a dictionary but can technically return anything.
    func entropy(filename string, r io.ReadSeeker, blocksize int) (interface{}, error) {
        var ret []float64
        ... /* compute Shannon's entropy on blocks of <blocksize> bytes */
        return map[string]interface{} { "entropy": ret }, nil
    }
    func main() {
        operators.AnalyzerRegister("entropy", entropy)
        m := molly.New()
        ...
    }

The new operator can be used in rules::

    rule example {
        ...
        analyze("entropy", "myentryopyanalisys", 1024);
    }


Custom extractors
------------------
Support for new extracttor formats can be added using the extractor API::

    func unpacker(e *types.Env, prefix string) (string, error) {
        inputfile := e.GetFile()
        r :=  e.Reader
        w, _, err := e.Create(prefix + inputfile + "_unpacked)
        // using e.Create() will ensure that the resulting file has a sane name and is fed
        // back into molly for analysis later

        ... /* unpacking from r to w using some custom algorithm */

        return "", nil
    }

    func main() {
        operators.ExtractorRegister("unpacker", unpacker)
        m := molly.New()
        ...
    }




Custom checksum functions
-------------------------

Molly provides a number of checksum functions such as *sha256* and *md5*. If you need something else, Molly provides API to add custom checksum functions::

    import (
        "crypto/md5"
        "hash"
        ...
    )

    func mysuperhash() hash.Hash {
        return md5.New
    }

    func main() {
        operators.RegisterChecksumFunction("mysuperhash", mysuperhash)
        m := molly.New()
        ...
    }

