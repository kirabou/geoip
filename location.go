
package geoip


// This package provides function to manage the GeoIP location file
// from MaxMind LLC.

import (
	"fmt"
	"os"
	"encoding/csv"
	"io"
	"bufio"
	"strconv"
)


// This the type to hold all information about a location.
// Country and Region are 2 characters string as defined in
// ISO 3661-1 alpha 2. 
// Example : 
// 	location[718]= { "US","MA","Medway","02053","42.1556","-71.4268","506","508" }
type Location struct {
	Country string	
	Region string
	City string
	PostalCode string
	Latitude string
	Longitude string
	MetroCode string
	AreaCode string
}


// Default path to MaxMind locations file
const LOCATIONS_FILE = "/tmp/GeoLiteCity-Location.csv"


var regions_tree *Regions
var countries_tree *Countries


// Returns country name of a given Location or ""
func (loc *Location)GetCountry() string {

	if countries_tree == nil {
		return ""
	}

	country := countries_tree.Get(loc.Country)
	if country == nil {
		return ""
	} else {
		return country.Name
	}

}


// Returns region name of a given location or ""
func (loc *Location)GetRegion() string {

	if regions_tree == nil {
		return ""
	}

	code := fmt.Sprintf("%s%s", loc.Country, loc.Region)

	region := regions_tree.Get(code)
	if region == nil {
		return ""
	} else {
		return region.Name
	}

}


// Implements String() function to Location type, so it
// implements the Stringer interface an can be Println()
func (loc *Location) String() string {
	country := loc.GetCountry()
	// fmt.Println("Country:", country)
	region := loc.GetRegion()
	// fmt.Println("Region:", region)
	return fmt.Sprintf("Country=%q (%s), Region=%q (%s), City=%q, PostalCode=%q, Latitude=%q, Longitude=%q, MetroCode=%q, AreaCode=%q",
			loc.Country, country, loc.Region, region, loc.City, loc.PostalCode, loc.Latitude, loc.Longitude, loc.MetroCode, loc.AreaCode)
}


// Returns number of lines in an io.Reader (like an
// open file)
func countLine(io io.Reader) int {
	fileScanner := bufio.NewScanner(io)
	lineCount := 0
	for fileScanner.Scan() {
    	lineCount++
	}
	log_geolocip.Notice(fmt.Sprintf("Locations number of lines: %d", lineCount))
	return lineCount
}


// Read a MaxMind GeoIP Location file in memory, as a
// slice of Location structures. For a known location_id,
// the location information will be found at Location[location_id].
func LoadLocFile(filename string) ([]Location, error) {
    
    file, err := os.Open(filename)
    if err != nil {
		log_geolocip.Err(fmt.Sprintf("Locations error open file: %v", err))
        return []Location{}, err
    }
    defer file.Close()

    // Build a slice big enough to hold all the locations
    line_count := countLine(file)
    loc_list := make([]Location, line_count)

    // Reset file position after counting the lines
    file.Seek(0, 0)

    // Use a CSV scanner to read file. Because the MaxMind files are
    // iso8859-1 encoded, we are using a fileLatin1Reader to convert
    // the read content to utf-8
    flr := fileLatin1Reader{ file: file }
    r := csv.NewReader(&flr)
    r.FieldsPerRecord = -1

    for {
    
    	values, err := r.Read()
    	if err == io.EOF {
    		break
    	}
    	if err != nil {
			log_geolocip.Err(fmt.Sprintf("Locations error reading file: %v", err))
    		break
    	}
	
		// Use only lines with 9 values
	   	if len(values) == 9 {

	   		locId, err := strconv.Atoi(values[0])
	   		if err != nil {
	   			// log.Println("Line ignored, cannot read LocId", err)
	   			continue
	   		}	   		

	   		loc_list[locId] = Location {
	   			Country: values[1],
	   			Region: values[2],
	   			City: values[3],
	   			PostalCode: values[4],
	   			Latitude: values[5],
	   			Longitude: values[6],
	   			MetroCode: values[7],
	   			AreaCode: values[8],
	   		}

	   	}
    }

    countries_tree, _ = LoadCountries()
    regions_tree, _ = LoadRegions()

    return loc_list, nil
}












