package cmd

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"time"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

type compression uint8

const (
	noCompression compression = iota
	deflateCompression
)

type localFileHeader struct {
	signature        uint32
	version          uint16
	bitflag          uint16
	compression      compression
	lastModified     time.Time
	crc32            uint32
	compressedSize   uint32
	uncompressedSize uint32
	fileName         string
	extraField       []byte
	fileContents     string
	fileContentsByte []byte
}

var errNotZip = fmt.Errorf("Not a zip file")
var errOverranBuffer = fmt.Errorf("Overran buffer")

var listCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		// read filename
		f, err := ioutil.ReadFile(fileName)
		check(err)

		end := 0
		for end < len(f) {
			lfh, next, err := parseLocalFileHeader(f, end)
			if err == errNotZip && end > 0 {
				break
			}
			check(err)

			end = next

			b1 := lfh.fileContentsByte[0]
			b2 := lfh.fileContentsByte[1]
			b3 := lfh.fileContentsByte[2]
			fmt.Printf("Byte 1: %#X %v\n", b1, b1)
			fmt.Printf("Byte 2: %#X %v\n", b2, b2)
			fmt.Printf("Byte 3: %#X %v\n", b3, b3)

			fmt.Println(lfh.lastModified, lfh.fileName, lfh.fileContents)
			fmt.Printf("%#v", lfh)
		}
	},
}

func parseLocalFileHeader(bs []byte, start int) (*localFileHeader, int, error) {

	// Check signature
	signature, i, err := readUint32(bs, start)
	if signature != 0x04034b50 {
		return nil, 0, errNotZip
	}

	version, i, err := readUint16(bs, i)
	if err != nil {
		return nil, 0, err
	}

	bitFlag, i, err := readUint16(bs, i)
	if err != nil {
		return nil, 0, err
	}

	compression := noCompression
	compressionRaw, i, err := readUint16(bs, i)
	if nil != err {
		return nil, 0, err
	}
	if 8 == compressionRaw {
		compression = deflateCompression
	}

	lmTime, i, err := readUint16(bs, i)
	if nil != err {
		return nil, 0, err
	}

	lmDate, i, err := readUint16(bs, i)
	if nil != err {
		return nil, 0, err
	}
	lastModified := msdosTimeToGoTime(lmDate, lmTime)

	crc32, i, err := readUint32(bs, i)
	if nil != err {
		return nil, 0, err
	}

	compressedSize, i, err := readUint32(bs, i)
	if nil != err {
		return nil, 0, err
	}

	uncompressedSize, i, err := readUint32(bs, i)
	if nil != err {
		return nil, 0, err
	}

	fileNameLength, i, err := readUint16(bs, i)
	if nil != err {
		return nil, 0, err
	}

	extraFieldLength, i, err := readUint16(bs, i)
	if nil != err {
		return nil, 0, err
	}

	fileName, _, i, err := readString(bs, i, int(fileNameLength))
	if nil != err {
		return nil, 0, err
	}

	extraField, i, err := readBytes(bs, i, int(extraFieldLength))
	if nil != err {
		return nil, 0, err
	}

	var fileContents string
	var fileContentsByte []byte
	if compression == noCompression {
		fileContents, fileContentsByte, i, err = readString(bs, i, int(uncompressedSize))
		if nil != err {
			return nil, 0, err
		}
	} else {
		end := i + int(compressedSize)
		if end > len(bs) {
			return nil, 0, errOverranBuffer
		}

		flateReader := flate.NewReader(bytes.NewReader(bs[i:end]))
		defer flateReader.Close()

		read, err := ioutil.ReadAll(flateReader)
		if nil != err {
			return nil, 0, err
		}

		fileContentsByte = read
		fileContents = string(read)
		i = end
	}

	return &localFileHeader{
		signature:        signature,
		version:          version,
		bitflag:          bitFlag,
		compression:      compression,
		lastModified:     lastModified,
		crc32:            crc32,
		compressedSize:   compressedSize,
		uncompressedSize: uncompressedSize,
		fileName:         fileName,
		extraField:       extraField,
		fileContents:     fileContents,
		fileContentsByte: fileContentsByte,
	}, i, nil
}

func readUint32(bs []byte, offset int) (uint32, int, error) {
	end := offset + 4
	if end > len(bs) {
		return 0, 0, errOverranBuffer
	}

	return binary.LittleEndian.Uint32(bs[offset:end]), end, nil
}

func readUint16(bs []byte, offset int) (uint16, int, error) {
	end := offset + 2
	if end > len(bs) {
		return 0, 0, errOverranBuffer
	}

	return binary.LittleEndian.Uint16(bs[offset:end]), end, nil
}

func readBytes(bs []byte, offset int, n int) ([]byte, int, error) {
	end := offset + n
	if end > len(bs) {
		return nil, 0, errOverranBuffer
	}

	return bs[offset : offset+n], end, nil
}

func readString(bs []byte, offset int, n int) (string, []byte, int, error) {
	read, end, err := readBytes(bs, offset, n)
	return string(read), read, end, err
}

func msdosTimeToGoTime(d uint16, t uint16) time.Time {
	seconds := int((t & 0x1F) * 2)
	minutes := int((t >> 5) & 0x3F)
	hours := int(t >> 11)

	day := int(d & 0x1F)
	month := time.Month((d >> 5) & 0x0F)
	year := int((d>>9)&0x7F) + 1980
	return time.Date(year, month, day, hours, minutes, seconds, 0, time.Local)
}

func check(err error) {
	if nil != err {
		panic(err)
	}
}
