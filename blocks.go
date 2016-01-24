
package geoip

import (
	"fmt"
	"os"
	"encoding/csv"
	"io"
	"strconv"
	"github.com/google/btree"
)


// This package provides function to manage the GeoIP Blocks file
// from MaxMind LLC.

// A Block is a range of IP addresses (from LowIP to
// HighIP) matching a given location ID (LocID). Blocks
// cannot overlap. Block example : 
// 	{ 16777216, 16777471, 17 }
type Block struct {
	LowIP uint32
	HighIP uint32
	LocId uint32
}


// All blocks are stored in memory in a BTree.
type Blocks btree.BTree


// Default filename for the MaxMind LLC blocks file
const BLOCKS_FILE = "/tmp/GeoLiteCity-Blocks.csv"


// Implements String() function to Block type, so it
// implements the Stringer interface an can be Println().
func (block *Block) String() string {
	return fmt.Sprintf("LowIP=%d, HighIP=%d, LocId=%d",
		block.LowIP, block.HighIP, block.LocId)
}


// Implements the Item interface from btree package for
// the Block type, so we can use them in a btree.
func (block Block)Less(than btree.Item) bool {

	// Less tests whether the current item is less than the given argument.
	return block.HighIP < than.(Block).LowIP

}



// Read a MaxMind GeoIP Blocks file in memory, as a
// BTree of Blocks structures.
func LoadBlocksFile(filename string) (*Blocks, error) {
    
    file, err := os.Open(filename)
    if err != nil {
    	log_geolocip.Err(fmt.Sprintf("Blocks error open file: %v", err))
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
    		log_geolocip.Err(fmt.Sprintf("Blocks error reading file: %v", err))
    		break
    	}
	
		// Use only lines with 3 values
	   	if len(values) == 3 {

	   		low_ip, err := strconv.ParseUint(values[0], 10, 32)
	   		if err != nil {
	   			// log.Println("Line ignored, cannot read LowIP", err)
	   			continue
	   		}	   		
	   		high_ip, err := strconv.ParseUint(values[1], 10, 32)
	   		if err != nil {
	   			// log.Println("Line ignored, cannot read HighIP", err)
	   			continue
	   		}	   		
	   		loc_id, err := strconv.ParseUint(values[2], 10, 32)
	   		if err != nil {
	   			// log.Println("Line ignored, cannot read LocId", err)
	   			continue
	   		}	   		

	   		var block = Block{ uint32(low_ip), uint32(high_ip), uint32(loc_id) }
	   		// fmt.Println(block)
	   		t.ReplaceOrInsert(block)

	   	}
    }

    return (*Blocks)(t), nil
}


// Returns the Block structure matching a given IP address.
func (blocks *Blocks)Get(IP uint32) *Block {
	tree := (*btree.BTree)(blocks)
	item := tree.Get(Block{IP, IP, 0}) // .(Block)
	if item != nil {
		block := item.(Block)
		return(&block)
	} else {
		return(nil)
	}
}

