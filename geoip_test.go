package geoip

import (
	"testing"
	"log"
	"net"
	"encoding/json"
	"os"
	"io"
)



func TestGeoLocIPv4(t *testing.T) {
	gli := GeoLocIPv4(net.ParseIP("54.88.55.63"))
	log.Println(gli)
	if gli == nil || gli.Location.City != "Ashburn" || *(gli.CountryName) != "États-Unis" || *(gli.RegionName) != "Virginia" {
		t.Errorf("Failed : geolocation for test IP does not match")	
	}
	if gli != nil {
		buf, _ := gli.MarshalJSON()
		log.Println(string(buf))
	}
	if gli != nil {
		json, _ := json.Marshal(gli)
		log.Println(string(json))
	}
}


func TestLatin1Reader(t *testing.T) {
	sample := []byte("\xc0\xc1\xc7\xc8\xc9ABCD\xca\xe0\xe1\xe2\xe7\xe8\xe9\xea\xee\xef\xf2\xf4\xf9\xfb\xff\xaeE") // latin1 for "ÀÁÇÈÉABCDÊàáâçèéêîïòôùûÿ®E"
	file, err := os.Create("/tmp/iso8859-1.txt")
	if err != nil {
		t.Errorf("Cannot create test file: %v", err)
	}
	_, err = file.Write(sample)
	if err != nil {
		t.Errorf("Cannot write test file: %v", err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Cannot close test file after writing: %v", err)
	}	

	file, err = os.Open("/tmp/iso8859-1.txt")
	if err != nil {
		t.Errorf("Cannot open file for reading: %v", err)
	}	

	flr := &fileLatin1Reader{ file: file }

	buf := make([]byte, 5)
	var read_sample []byte
	for n, err := flr.Read(buf); n>0; n, err = flr.Read(buf) {
		// First handle the read bytes
		read_sample = append(read_sample, buf[:n]...)

		// Next, handle error conditions
		if err != nil {
			if err != io.EOF {
				t.Errorf("Error reading test file: %v", err)
			}
			break;
		}
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Cannot close test file after reading: %v", err)
	}	
	if string(read_sample) != "ÀÁÇÈÉABCDÊàáâçèéêîïòôùûÿ®E" {
		t.Errorf("Converted string does not match sample")
	}

}
