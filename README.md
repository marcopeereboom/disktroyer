# disktroyer
file system test tool

disktroyer creates a number of working directories and fills them with files
containing random data.  Once all files are created they are all moved to
another directory.  Once moved the files are deleted.  This is repeated until
the user aborts the test with ctrl-c.

All this fun happens in parallel.

#install
Install go then:
```
go get github.com/marcopeereboom/disktroyer
```

# usage
Starting disktroyer without parameters will launch io relative to the current
directory.

```
  Usage of disktroyer:
    -debug=false: enable golang pprof
    -maxdirs=16: number of working directories
    -maxfiles=100: maximum number of files per directory
    -maxfilesize=65536: maximum file size
    -root="disktroyer": root directory for test
    -verbose=false: enable verbosity
```

# disclaimer
If your disk or computer goes up in a ball of flames please record it and
share.  We'd love to see it!

Oh, and I don't care so don't blame me.
