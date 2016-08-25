package getfiles

import (
	"compress/gzip"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/Clever/pathio.v3"

	"github.com/Clever/http-science/config"
)

// AddFilesToChan adds files from the specified location to a chan
func AddFilesToChan(payload *config.Payload, files chan<- string) error {
	base := "%s"
	if payload.S3Bucket != "" {
		base = fmt.Sprintf("s3://%s/%%s", payload.S3Bucket)
	} else {
		return fmt.Errorf("Currently only supports s3 files")
	}

	baseWithPrefix := fmt.Sprintf(base, payload.FilePrefix)
	if payload.FilePrefix != "" {
		baseWithPrefix += "/"
	}

	// Starting with the baseWithPrefix, build a stack of directories to explore and
	// files to download.
	fileStack := []string{baseWithPrefix}
	for len(fileStack) > 0 {
		file := fileStack[len(fileStack)-1]
		fileStack = fileStack[:len(fileStack)-1]

		fileType, err := getFileType(file, baseWithPrefix)
		if err != nil {
			return err
		}

		if fileType == "file" {
			localfile, err := downloadFile(file)
			if err != nil {
				return fmt.Errorf("s3 download failed: %s", err)
			}
			files <- localfile
		} else {
			newFiles, err := goDeeper(file, fileType, base, baseWithPrefix, payload)
			if err != nil {
				return err
			}
			fileStack = append(fileStack, newFiles...)
		}
	}
	return nil
}

// downloadFile downloads, unzips and writes to /tmp/filename.txt
func downloadFile(file string) (string, error) {
	reader, err := pathio.Reader(file)
	defer reader.Close()

	if err != nil {
		return "", err
	}
	decompress, err := gzip.NewReader(reader)
	defer decompress.Close()
	if err != nil {
		return "", err
	}

	finalRes := []byte{}
	chunk := 10000 // read in 10KB chunks
	for {
		res := make([]byte, chunk)
		n, err := decompress.Read(res)
		if err != nil {
			return "", err
		}
		finalRes = append(finalRes, res[:n]...)
		if n != chunk {
			break
		}
	}
	filename := fmt.Sprintf("%s/%s.txt", os.TempDir(), strings.Split(finalPath(file), ".")[0])
	pathio.Write(filename, finalRes)
	return filename, nil
}

// NextType maps a file type to the type that comes after it
var NextType = map[string]string{
	"base":  "year",
	"year":  "month",
	"month": "day",
	"day":   "hour",
	"hour":  "file",
}

// getFileType looks at the filename and determines what type it is
func getFileType(file, baseWithPrefix string) (string, error) {
	typeRegex := map[string]string{
		"file":  "^" + baseWithPrefix + "[0-9]{4}/[0-9]{2}/[0-9]{2}/[0-9]{2}/.+$",
		"hour":  "^" + baseWithPrefix + "[0-9]{4}/[0-9]{2}/[0-9]{2}/[0-9]{2}/$",
		"day":   "^" + baseWithPrefix + "[0-9]{4}/[0-9]{2}/[0-9]{2}/$",
		"month": "^" + baseWithPrefix + "[0-9]{4}/[0-9]{2}/$",
		"year":  "^" + baseWithPrefix + "[0-9]{4}/$",
		"base":  "^" + baseWithPrefix + "$",
	}

	for t, regex := range typeRegex {
		match, err := regexp.MatchString(regex, file)
		if err != nil {
			return "", err
		}
		if match {
			return t, nil
		}
	}
	return "", fmt.Errorf("Type not found for %s", file)
}

// goDeeper returns a list of files in directory file that satisfy:
// Are of the correct format (needed because of https://luceeserver.atlassian.net/browse/LDEV-359)
// Are before the start_before param
// Are the right file for this job_number
func goDeeper(file, fileType, base, baseWithPrefix string, payload *config.Payload) ([]string, error) {
	newFiles, err := pathio.ListFiles(file)
	if err != nil {
		return nil, err
	}

	filesToUse := []string{}
	for i := range newFiles {
		fullPath := fmt.Sprintf(base, newFiles[i])
		nextType, err := getFileType(fullPath, baseWithPrefix)

		// % totalJobs to handle multiple files from one directory
		// -1 because mod goes from 0 and JobNumber from 1
		// % both by len(newFiles) to handle (JobNumber > len(newFiles)) which would result in high numbered
		// jobs getting no files
		forThisJob := i%payload.TotalJobs%len(newFiles) == ((payload.JobNumber - 1) % len(newFiles))

		if err != nil {
			return nil, err
		} else if nextType != NextType[fileType] { // ignore files that don't match the expected regex
			continue
		} else if tooRecent(fullPath, nextType, payload) { // ignore files after the startBefore date
			continue
		} else if nextType == "file" && !forThisJob {
			continue
		}

		filesToUse = append(filesToUse, fullPath)
	}
	return filesToUse, nil
}

// tooRecent returns false if the file's date is after the start_before param else true
func tooRecent(file, fileType string, payload *config.Payload) bool {
	switch fileType {
	case "year":
		return finalPath(file) > payload.StartBefore[:4]
	case "month":
		return finalPath(file) > payload.StartBefore[5:7]
	case "day":
		return finalPath(file) > payload.StartBefore[8:10]
	case "hour":
		return finalPath(file) > payload.StartBefore[11:13]
	case "file":
		return false
	}
	return true
}

// finalPath given a/b/c/d/ or a/b/c/d returns d
func finalPath(path string) string {
	res := strings.Split(path, "/")

	if path[len(path)-1] == '/' {
		return res[len(res)-2]
	}
	return res[len(res)-1]
}
