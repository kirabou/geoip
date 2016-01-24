
package geoip

import (
	"fmt"
	"io"
	"os"
	"encoding/csv"
	"strconv"
	"github.com/google/btree"
)


// This package provides function to manage the ASN file
// from MaxMind LLC.

// Default filename for the ASN file from MaxMind LLC
const ASN_FILE = "/tmp/GeoIPASNum2.csv"


// An ASN structure is a range of IP addresses (from LowIP
// to HighIP) matching a given ASN information string. 
// ASN example : 
// 	{ 16777216, 16777471, "AS15169 Google Inc." }
type ASN struct {
	LowIP uint32
	HighIP uint32
	ASN string
}


// All ASNs are kept in memory as a BTree. ASNs is the
// type for this btree.
type ASNs btree.BTree


// Implements String() function to *ASN type, so it
// implements the Stringer interface an can be Println().
func (asn *ASN) String() string {
	return fmt.Sprintf("LowIP=%d, HighIP=%d, ASN=%q",
		asn.LowIP, asn.HighIP, asn.ASN)
}


// Implements the Item interface from btree package for
// the ASN type, so we can use them in a btree.
func (asn ASN)Less(than btree.Item) bool {

	// Less tests whether the current item is less than the given argument.
	return asn.HighIP < than.(ASN).LowIP

}


// Read a MaxMind GeoIP ASN file in memory, as a BTree
// of ASN structures.
func LoadASNFile(filename string) (*ASNs, error) {
    
    file, err := os.Open(filename)
    if err != nil {
        log_geolocip.Err(fmt.Sprintf("ASN error open file: %v", err))
        return nil, err
    }
    defer file.Close()

    t := btree.New(4)

    r := csv.NewReader(file)
    r.FieldsPerRecord = -1

    for {
    
    	values, err := r.Read()
    	if err == io.EOF {
    		break
    	}    	
    	if err != nil {
    		log_geolocip.Err(fmt.Sprintf("ASN error reading file: %v", err))
    		break
    	}
	
		// Use only lines with 3 values
	   	if len(values) == 3 {

	   		low_ip, err := strconv.ParseUint(values[0], 10, 32)
	   		if err != nil {
	   			// fmt.Println("Line ignored, cannot read LowIP", err)
	   			continue
	   		}	   		
	   		high_ip, err := strconv.ParseUint(values[1], 10, 32)
	   		if err != nil {
	   			// fmt.Println("Line ignored, cannot read HighIP", err)
	   			continue
	   		}	   		

	   		// var asn = ASN{ uint32(low_ip), uint32(high_ip), values[2] }
	   		// fmt.Println(&asn)
	   		t.ReplaceOrInsert(ASN{ uint32(low_ip), uint32(high_ip), values[2] })

	   	}
    }

    return (*ASNs)(t), nil
}


// Returns ASN structure matching a given IP address.
func (asns *ASNs)Get(IP uint32) *ASN {
	tree := (*btree.BTree)(asns)
	item := tree.Get(ASN{IP, IP, ""})
	if item != nil {
		asn := item.(ASN)
		return(&asn)
	} else {
		return(nil)
	}
}

