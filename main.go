package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

func main() {
	// "62.141.54.25" - heisse.de
	// "142.1.217.155" - ixmaps.ca
	// "140.82.114.3" - github.com
	ips, _ := net.LookupIP("ixmaps.ca")            // get []net.IP for hostname string
	netip := net.ParseIP("142.1.217.155")          // get net.IP for IP string. Can be used to decide btw ip and hostname, returns nil eg for a hostname
	hostnames, _ := net.LookupAddr("140.82.114.3") // get []string hostnames for IP string
	localIPAny := getLocalIP()
	localIPExternal := getLocalIPExternal()

	fmt.Printf("LookupIP: %+v\n", ips)
	fmt.Printf("ParseIP: %+v\n", netip)
	fmt.Printf("Hostnames: %+v\n", hostnames)
	fmt.Printf("Local any: %v\n", localIPAny)
	fmt.Printf("Local external: %v", localIPExternal)

}

const URIPattern string = `^((ftp|http|https):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(((([a-z\x{00a1}-\x{ffff}0-9]+-?-?_?)*[a-z\x{00a1}-\x{ffff}0-9]+)\.)?)?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?_?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?)|localhost)(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`

func validateURI(uri string) bool {
	pattern := URIPattern
	match, err := regexp.MatchString(pattern, uri)
	check(err)
	return match
}

func getLocalIPExternal() string {
	req, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		check(err)
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		check(err)
	}
	return string(body)
}

// GetLocalIP returns the non loopback local IP of the host
func getLocalIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !isPrivateIP(ipnet.IP) {
			return ipnet.IP
			// if ipnet.IP.To4() != nil {
			// 	return ipnet.IP
			// }
		}
	}
	return nil
}

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
