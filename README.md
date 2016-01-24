# Introduction
geoip is a Go package that provides geoip informations for an IP address, based on MaxMind GeoIP files, and a REST API, inspired from Telize.com, to get geoip informations as a JSON structure.

All data are stored in memory for maximum speed. MaxMind files are automatically downloaded if the current files are older than 8 days. Initialization is made through `init()` and could take up to 30 seconds depending of your hardare configuration. Around 500MB of memory are required to store all geoip data.


# Most useful functions 

- `GeoLocIPv4()` returns a GeoLocIp structure for a given IPv4 address.

- `ServeHttpRequest()` provides a REST API, returning a JSON structure holding the geolocation information for a given IPv4 address.

- `ServeGeoLocAPI()` starts a dedicated http server that only provides the REST API.

- `MarshalJSON()` implements the JSON Marshaler interface for the `*GeoLocIp` type.


# Contact

Email to : kirabou (at) gmx.com


# Logging

Error and information messages are written to the local system log (syslog).


# Known limitations

- Currently works with IPv4 adresses only.

- Need to be restarted to reload GeoIP files from Maxmind.


# Licence

Distributed under the MIT licence.


# Acknowledgments 

- GeoIP data provided by MaxMind LLC.

- goggle/btree package used to store and search for data in memory.


# Examples

The following example starts an http server, listening on the 9001 port. It returns a JSON structure with all the geoip informations. For example :`http://localhost:9001/54.88.55.63` returns the following JSON structure :

```
    { 
        "ip":"54.88.55.63",
        "country_code":"US",
        "region_code":"VA",
        "city":"Ashburn",
        "postal_code":"20147",
        "latitude":39.0335,
        "longitude":-77.4838,
        "metro_code":"511",
        "area_code":"703",
        "organization":"AS14618 Amazon.com, Inc.",
        "country":"États-Unis",
        "region":"Virginia" 
    }
```

Here the source code  :

```
    package main
    
    import (
        "geoip"
    )
    
    func main() {
        geolocip.ServeGeoLocAPI(9001)
    }
```

The next example is a simple forever loop, waiting for an IPv4 address, and returning the geoip informations for it.

```
	package main

	import (
	    "geoip"
	)

	func main() {

	    for {

			fmt.Print("Enter IPv4 address in a.b.c.d format : ")
			var ip_address string
	    	fmt.Scanf("%s", &ip_address)
			ip := net.ParseIP(ip_address)
			if ip == nil {
			    fmt.Println("Not a valid IP address.")
			    continue
			}

			fmt.Println(ip)
			json, _ := json.Marshal(geolocip.GeoLocIPv4(ip))

			fmt.Println(string(json))

	    }
	    
	}
```
