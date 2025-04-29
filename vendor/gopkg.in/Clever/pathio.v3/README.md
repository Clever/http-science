# pathio

[![GoDoc](https://godoc.org/gopkg.in/Clever/pathio.v3?status.svg)](https://godoc.org/gopkg.in/Clever/pathio.v3)

--
    import "gopkg.in/Clever/pathio.v3"

Package pathio is a package that allows writing to and reading from different
types of paths transparently. It supports two types of paths:

    1. Local file paths
    2. S3 File Paths (s3://bucket/key)

Note that using s3 paths requires setting two environment variables

    1. AWS_SECRET_ACCESS_KEY
    2. AWS_ACCESS_KEY_ID

## Usage
Pathio has a very easy to use interface, with 4 main functions:

```
import "gopkg.in/Clever/pathio.v3"
var err error
```

### ListFiles
```
// func ListFiles(path string) ([]string, error)
files, err = pathio.ListFiles("s3://bucket/my/key") // s3
files, err = pathio.ListFiles("/home/me")           // local
```

### Write / WriteReader
```
// func Write(path string, input []byte) error
toWrite := []byte("hello world\n")
err = pathio.Write("s3://bucket/my/key", toWrite)   // s3
err = pathio.Write("/home/me/hello_world", toWrite) // local

// func WriteReader(path string, input io.ReadSeeker) error
toWriteReader, err := os.Open("test.txt") // this implements Write and Seek
err = pathio.WriteReader("s3://bucket/my/key", toWriteReader)   // s3
err = pathio.WriteReader("/home/me/hello_world", toWriteReader) // local
```

### Read
```
// func Reader(path string) (rc io.ReadCloser, err error)
reader, err = pathio.Reader("s3://bucket/key/to/read") // s3
reader, err = pathio.Reader("/home/me/file/to/read")   // local
```
