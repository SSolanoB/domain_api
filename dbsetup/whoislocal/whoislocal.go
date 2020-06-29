package whoislocal

import (
	"fmt"
	"github.com/likexian/whois-go"
	
	"regexp"
	"strings"
)

func AskforIp(query string) (country, org_name string, err error){
	//query := "2607:f8b0:4005:808:0:0:0:200e" ipv6
	//query = "205.251.242.103" ipv4
	request, err := whois.Whois(query)
	if err == nil {
		fmt.Println(request)
		//result, err := whoisparser.Parse(request)
		re := regexp.MustCompile(`(?:[Cc]ountry:) (.*)`)
		country := re.FindAllStringSubmatch(string(request), -1)[0][1]
		fmt.Printf("%q\n", strings.TrimSpace(country))

		re_2 := regexp.MustCompile(`(?:[Oo]rg-?[Nn]ame:) (.*)`)
		org_name := re_2.FindAllStringSubmatch(string(request), -1)[0][1]
		fmt.Printf("%q\n", strings.TrimSpace(org_name))
		return country, org_name, err
	} else {
		var country string
		var org_name string
		return country, org_name, err
	}

}