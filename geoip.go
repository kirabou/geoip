// Provides geoip information for an IP address, based on MaxMind GeoIP files, and 
// a REST API, inspired from Telize.com, to get geoip information as a JSON
// structure.
// 
// All data are stored in memory for maximum speed. MaxMind files are automatically
// downloaded if the current files are older than 8 days. Initialization is made
// through init() and could take up to 30 seconds depending of your hardware configuration.
// Around 500MB of memory are required to store all geoip data.
// 
// 
// Most useful functions 
// 
// GeoLocIPv4() returns a GeoLocIp structure for a given IPv4 address.
// 
// ServeHttpRequest() provides a REST API, returning a JSON structure holding
// the geolocation information for a given IPv4 address.
// 
// ServeGeoLocAPI() starts a dedicated http server that only provides the REST API.
// 
// MarshalJSON() implements the JSON Marshaler interface for the *GeoLocIp
// type.
// 
// 
// Contact
// 
// Email to: kirabou (at) gmx.com
// 
// 
// Logging
// 
// Error and information messages are written to the local system log (syslog).
// 
// 
// Known limitations
// 
// Currently works with IPv4 addresses only.
// 
// Need to be restarted to reload GeoIP files from MaxMind.
// 
// 
// License
// 
// Distributed under the MIT license.
// 
// 
// Acknowledgments 
// 
// GeoIP data provided by MaxMind LLC.
// 
// goggle/btree package used to store and search for data in memory.
// 
// 
// Installing and testing
//
// The package can be installed using the following command : 
//   go get github.com/kirabu/geoip 
// It will install both the geoip package 
// and the google/btree package.
//
// The tests can be runned with
//   go test github.com/kirabu/geoip 
// If you check your system log (/var/log/syslog), you'll see the main
// steps followed by geoip to download the MaxMind files and load
// them into memory :
// 
//   Jan 25 05:43:35  geolocip[4204]: Starting
//   Jan 25 05:43:35  geolocip[4204]: Download http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum2.zip
//   Jan 25 05:43:37  geolocip[4204]: Extracted /tmp/GeoIPASNum2.csv
//   Jan 25 05:43:37  geolocip[4204]: Download http://geolite.maxmind.com/download/geoip/database/GeoLiteCity_CSV/GeoLiteCity-latest.zip
//   Jan 25 05:43:50  geolocip[4204]: Extracted /tmp/GeoLiteCity-Blocks.csv
//   Jan 25 05:43:51  geolocip[4204]: Extracted /tmp/GeoLiteCity-Location.csv
//   Jan 25 05:43:51  geolocip[4204]: Locations number of lines: 751379
//   Jan 25 05:43:55  geolocip[4204]: Locations file loaded
//   Jan 25 05:44:04  geolocip[4204]: Blocks file loaded
//   Jan 25 05:44:05  geolocip[4204]: ASN file loaded
// 
// 
// Examples
// 
// The following example starts an http server, listening on the 9001
// port. It returns a JSON structure with all the geoip information.
// For example :
// 	http://localhost:9001/54.88.55.63
// returns the following JSON structure :
// 	{ "ip":"54.88.55.63",
// 	  "country_code":"US",
// 	  "region_code":"VA",
// 	  "city":"Ashburn",
// 	  "postal_code":"20147",
// 	  "latitude":39.0335,
// 	  "longitude":-77.4838,
// 	  "metro_code":"511",
// 	  "area_code":"703",
// 	  "organization":"AS14618 Amazon.com, Inc.",
// 	  "country":"États-Unis",
// 	  "region":"Virginia" }
// 
// Here the source code  :
// 
// 	package main
//
//	import (
// 	    "geoip"
// 	)
//
// 	func main() {
// 	    geolocip.ServeGeoLocAPI(9001)
// 	}
// 
// The next example is a simple forever loop, waiting for an IPv4 address, and returning
// the geoip information for it.
// 	package main
//
// 	import (
//	    "geoip"
//	)
//
// 	func main() {
// 	
// 	    for {
//
//		fmt.Print("Enter IPv4 address in a.b.c.d format : ")
//		var ip_address string
//		    fmt.Scanf("%s", &ip_address)
//
//		ip := net.ParseIP(ip_address)
//		if ip == nil {
//		    fmt.Println("Not a valid IP address.")
//		    continue
//		}
//		
//		fmt.Println(ip)
//		json, _ := json.Marshal(geolocip.GeoLocIPv4(ip))
//		
//		fmt.Println(string(json))
//		
// 	    }
// 	    
// 	}
package geoip


import (
	"log"
	"fmt"
	"net"
	"bytes"
	"bufio"
	"encoding/json"
	"net/http"
	"path"
	"log/syslog"
	"os"
	"io"
	"archive/zip"
	"errors"
	"time"
)


var locations []Location
var blocks *Blocks
var asn_tree *ASNs
var log_geolocip *syslog.Writer


// This is the structure type used to share
// geolocation information for a given IP
type GeoLocIp struct {
	Ip net.IP 				
	Block *Block 			
	Location *Location
	Asn *ASN
	CountryName *string
	RegionName *string
}



// Implements String() function to *GeoLocIp type, so it
// implements the Stringer interface an can be Println()
func (gli *GeoLocIp) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, CountryName=%q, RegionName=%q",
		gli.Ip.String(), 
		fmt.Sprintf("%s", gli.Block),
		fmt.Sprintf("%s", gli.Location),
		fmt.Sprintf("%s", gli.Asn),
		*(gli.CountryName), *(gli.RegionName))
}



// Implements the json.Marshaler interface for the GeoLocIp, so it can
// be used with the standard decoding functions from the json package.
// Example of returned JSON for 54.88.55.63 :
//  {
//  	"ip":"54.88.55.63",
//  	"country_code":"US",
//  	"region_code":"VA",
//  	"city":"Ashburn",
//  	"postal_code":"20147",
//  	"latitude":39.0335,
//  	"longitude":-77.4838,
//  	"metro_code":"511",
//  	"area_code":"703",
//  	"organization":"AS14618 Amazon.com, Inc.",
//  	"country":"États-Unis",
//  	"region":"Virginia"
//  }
//  
// Not all fields are present, depending of available data.
func (gli *GeoLocIp) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
    w := bufio.NewWriter(&b)

    fmt.Fprintf(w, "{ \"ip\":%q", gli.Ip.String())

    if gli.Location != nil {
	    if gli.Location.Country != "" {
		    fmt.Fprintf(w, ", \"country_code\":%q", gli.Location.Country)
		}
	    if gli.Location.Region != "" {
		    fmt.Fprintf(w, ", \"region_code\":%q", gli.Location.Region)
		}
	    if gli.Location.City != "" {
	    	if tmp, err := json.Marshal(gli.Location.City); err == nil {
		    	fmt.Fprintf(w, ", \"city\":%s", tmp)
		    }
		}	
	    if gli.Location.PostalCode != "" {
	    	if tmp, err := json.Marshal(gli.Location.PostalCode); err == nil {
			    fmt.Fprintf(w, ", \"postal_code\":%s", tmp)
			}
		}	
	    if gli.Location.Latitude != "" {
		    fmt.Fprintf(w, ", \"latitude\":%s", gli.Location.Latitude)
		}	
	    if gli.Location.Longitude != "" {
		    fmt.Fprintf(w, ", \"longitude\":%s", gli.Location.Longitude)
		}	
	    if gli.Location.MetroCode != "" {
	    	if tmp, err := json.Marshal(gli.Location.MetroCode); err == nil {
			    fmt.Fprintf(w, ", \"metro_code\":%s", tmp)
			}
		}	
	    if gli.Location.AreaCode != "" {
		    if tmp, err := json.Marshal(gli.Location.AreaCode); err == nil {
		   		fmt.Fprintf(w, ", \"area_code\":%s", tmp)
		   	}
		}	
	}
    if gli.Asn != nil && gli.Asn.ASN != "" {
    	if tmp, err := json.Marshal(gli.Asn.ASN); err == nil {
	    	fmt.Fprintf(w, ", \"organization\":%s", tmp)
	    }
	}	
	if *(gli.CountryName) != "" {
		if tmp, err := json.Marshal(*(gli.CountryName)); err == nil {
	   		fmt.Fprintf(w, ", \"country\":%s", tmp)
	   	}
	}
	if *(gli.RegionName) != "" {
		if tmp, err := json.Marshal(*(gli.RegionName)); err == nil {
	    	fmt.Fprintf(w, ", \"region\":%s", tmp)
	    }
	}

	fmt.Fprintf(w, " }\n")
	w.Flush()

	return b.Bytes(), nil
}


// Loads blocks, locations, ASN, countries and regions in memory
func init() {

	var err error

	log_geolocip, err = syslog.New(syslog.LOG_NOTICE, "geolocip")
	// What should we do if syslog.New() returns an error ??
	if err != nil {
			log.Println("Cannot open log :", err)
			return
		}

	log_geolocip.Notice("Starting")

	DownloadMaxmindFiles()

	if locations == nil {
		locations, err = LoadLocFile(LOCATIONS_FILE)
		if err != nil {
			log_geolocip.Err(fmt.Sprintf("Cannot load locations file : %v", err))
			return
		}
	}
	log_geolocip.Notice("Locations file loaded")


	if blocks == nil {
		blocks, err = LoadBlocksFile(BLOCKS_FILE)
		if err != nil {
			log_geolocip.Err(fmt.Sprintf("Cannot load blocks file : %v", err))
			return
		}
	}
	log_geolocip.Notice("Blocks file loaded")

	if asn_tree == nil {
		asn_tree, err = LoadASNFile(ASN_FILE)
		if err != nil {
			log_geolocip.Err(fmt.Sprintf("Cannot load ASN file : %v", err))
			return
		}
	}
	log_geolocip.Notice("ASN file loaded")

}

// Returns the geolocation information for a given IPv4 address
// aa a *GeoLocIP if found, or nil
func GeoLocIPv4(ip net.IP) *GeoLocIp {

	if locations == nil || blocks == nil || asn_tree == nil {
		log_geolocip.Err("geoloip package badly initialized")
		return nil
	}

	addr := uint32(ip[15])+256*(uint32(ip[14])+256*(uint32(ip[13])+256*uint32(ip[12])))

	block := blocks.Get(addr)
   	if block == nil {
   		log_geolocip.Notice(fmt.Sprintf("No block found for IP %d %s", addr, ip.String()))
   		return nil
   	}

   	location := &locations[block.LocId]
   	country := location.GetCountry()
   	region := location.GetRegion()

   	return &(GeoLocIp{ip, block, location, asn_tree.Get(addr), &country, &region})

}


//  This serves an http request and returns the GeoLocIp information 
//  as a JSON for the IP address given in the URL path. See ServeGeoLocAPI()
//  and MarshalJSON(). If no IP address is given in the URL, this function
//  will try to use the IP of the caller.
func ServeHttpRequest(writer http.ResponseWriter, request *http.Request) {
	base := path.Base(request.URL.Path)
	var ip net.IP
	if base == "/" {
		host, _, _ := net.SplitHostPort(request.RemoteAddr)
		if host != "" {
			ip = net.ParseIP(host)
		}
	} else {
		ip = net.ParseIP(path.Base(request.URL.Path))
	}
	if ip != nil {
		json, _ := json.Marshal(GeoLocIPv4(ip))
		fmt.Fprintf(writer, string(json))
	}
}


// Starts an HTTP server on a local port whose number is given as argument. 
// It will serve requests for geolocation information of IP addresses. 
// For example : "http:your_host/54.88.55.63".
// See ServeHttpRequest() for a description of the returned JSON.
func ServeGeoLocAPI(port uint16) {
	http.HandleFunc("/", ServeHttpRequest)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
   		log_geolocip.Err(fmt.Sprintf("Cannot start http server: %v", err))
    }
}


// Download Maxmind files in /tmp
func download(url string, filename string) error {

	out, err := os.Create(filename)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Cannot create %s: %v", filename, err))
		return err
	}
	defer out.Close()

	in, err := http.Get(url)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Cannot get URL %s: %v", url, err))
		return err
	}
	defer in.Body.Close()

	_, err = io.Copy(out, in.Body)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Error downloading %s from %s: %v", filename, url, err))
		return err
	}

	return nil

}


// Returns age of a given file in days, or -1 if not found
// or error
func ageFile(filename string) int {
	fi, err := os.Stat(filename)
	if err != nil {
		return -1
	}
	return int(time.Since(fi.ModTime()).Hours()) / 24
}


// Extract file from a zip archive to a given filename
func extractFile(in_file *zip.File, out_file string) error {
	out, err := os.Create(out_file)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Cannot create %s: %v", out_file, err))
		return err
	}
	defer out.Close()
	in, err := in_file.Open()
    if err != nil {
		log_geolocip.Err(fmt.Sprintf("Cannot open archive for reading: %v", err))
		return err
    }	
    defer in.Close()
    io.Copy(out, in)
	log_geolocip.Notice(fmt.Sprintf("Extracted %s", out_file))
	return nil
}


// Pathname, name and URL for the Maxmind files
const (
	url_zipfile_asn = "http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum2.zip"
	url_zipfile_city = "http://geolite.maxmind.com/download/geoip/database/GeoLiteCity_CSV/GeoLiteCity-latest.zip"
	zipfile_asn = "/tmp/GeoIPASNum2.zip"
	zipfile_city = "/tmp/GeoLiteCity-latest.zip"
	file_asn = "GeoIPASNum2.csv"
	file_blocks = "GeoLiteCity-Blocks.csv"
	file_location = "GeoLiteCity-Location.csv"
)


// Download the Maxmind zip files if the current ones are older
// than 8 days. Extract files from the downloaded zip files.
func DownloadMaxmindFiles() error {
	// err := download(url_zipfile_city, zipfile_city)

	// ASN : check if file exists and is less than 8 days
	age_asn := ageFile(zipfile_asn)
	if age_asn == -1 || age_asn >= 8 {
		log_geolocip.Notice(fmt.Sprintf("Download %s", url_zipfile_asn))
		err := download(url_zipfile_asn, zipfile_asn)
		if err != nil {
			return err
		}	
	} else {
		log_geolocip.Notice(fmt.Sprintf("%s is %d days old", zipfile_asn, age_asn))
	}

	asn_zip, err := zip.OpenReader(zipfile_asn)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Error opening zip file %s: %v", zipfile_asn, err))
		return err
	} 
	defer asn_zip.Close()
	if asn_zip.File[0].Name != file_asn {
		log_geolocip.Err(fmt.Sprintf("Bad content in %s, found %s, expected %s", zipfile_asn, asn_zip.File[0].Name, file_asn))
		return errors.New("Bad content")		
	}

	if extractFile(asn_zip.File[0], ASN_FILE) != nil {
		return errors.New("Cannot extract ASN file")
	}

	// City : check if file exists and is less than 8 days
	age_city := ageFile(zipfile_city)
	if age_city == -1 || age_city >= 8 {
		log_geolocip.Notice(fmt.Sprintf("Download %s", url_zipfile_city))
		err := download(url_zipfile_city, zipfile_city)
		if err != nil {
			return err
		}	
	} else {
		log_geolocip.Notice(fmt.Sprintf("%s is %d days old", zipfile_city, age_city))
	}

	city_zip, err := zip.OpenReader(zipfile_city)
	if err != nil {
		log_geolocip.Err(fmt.Sprintf("Error opening zip file %s: %v", zipfile_city, err))
		return err
	} 
	defer city_zip.Close()
	for _, f := range city_zip.File {
		switch path.Base(f.Name) {
		case file_blocks :
			if extractFile(f, BLOCKS_FILE) != nil {
				return errors.New("Cannot extract Blocks file")
			}

		case file_location :
			if extractFile(f, LOCATIONS_FILE) != nil {
				return errors.New("Cannot extract Locations file")
			}
		}
	}

	return nil
}












