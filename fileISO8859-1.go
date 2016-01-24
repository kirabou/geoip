
package geoip



import (
	"os"
	"io"
)


// Implements the Reader interface for files with iso8859-1 (latin 1)
// contents. Read content is converted to utf-8. Latin-1 characters
// must be between 0x00 and 0xFF. Characters above 0x80 are converted
// to a 2 bytes utf-8 sequence. For more explanations see
// http://stackoverflow.com/questions/5586214/how-to-convert-char-from-iso-8859-1-to-utf-8-in-c-multiplatformly


// Note : the MaxMind files are iso8859-1 encoded (latin1).
// When parsing with the csv package, content must be utf8.
// Invalid utf8 character are replaced by U+FFFD. 
// When that character gets encoded back to UTF-8, this results in the byte sequence EF BF BD


type fileLatin1Reader struct {
	file *os.File  			// The file used to read data
	currentChar byte 		// In case we did not have enough space to write a 2 bytes utf-8
							// char into the caller's buffer, we store the 2nd byte here for
							// later use.
}

// Implements the reader interface, so we could have a reader
// that is able to convert iso8859-1 (latin1) to utf-8
func (flr *fileLatin1Reader)Read(p []byte) (n int, err error) {

	// Make a buffer the same size as p, and read file
	// content into this new buffer
	var buf []byte = make([]byte, len(p))
	n, err = flr.file.Read(buf)

	// Put the read content into the caller's buffer, while
	// converting it to utf-8.
	
	nb_written := 0 			// Number of written bytes in caller's buffer
	var i int 					// Number of transfered bytes from our own buffer

	// Start with previous unfinished utf-8 sequence, if any
	if flr.currentChar != 0 {
		p[nb_written] = flr.currentChar
		nb_written++
		flr.currentChar = 0
	}

	// Move bytes from our buffer, and convert to utf-8
	for i = 0; i<n; i++ {
		if nb_written >= len(p) {
			break
		}
		if buf[i] < 0x80 {
			p[nb_written] = buf[i]
			nb_written++
		} else {
			p[nb_written] = 0xC0 | (buf[i] & 0xC0) >> 6 
			nb_written++
			if nb_written >= len(p) {
				flr.currentChar = 0x80 | (buf[i] & 0x3f)
			} else {
				p[nb_written] = 0x80 | (buf[i] & 0x3f)
				nb_written++
			} 
		}
	}

	// Now i holds the actual number of bytes transfered from the file
	// to the caller's buffer, which can be less than n, because of the
	// possible utf-8 sequences added while transfering the bytes. So we
	// need to set the file position back to where the next read should
	// start.
	flr.file.Seek(int64(i-n), 1)
	
	// Special case : we may have reached the file EOF, but due to utf-8 sequences 
	// added to the bytes stream, we have not yet finished to transfer the
	// converted bytes
	if err == io.EOF && flr.currentChar != 0 {
		return nb_written, nil
	} else {	
		return nb_written, err
	}
}

