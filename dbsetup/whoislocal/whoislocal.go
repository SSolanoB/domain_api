package whoislocal

import (
	//"log"
	"fmt"
	"github.com/likexian/whois-go"
	//"github.com/likexian/whois-parser-go"
	"regexp"
	"strings"
)

func AskforIp(query string) (country, org_name string, err error){
	//query := "2607:f8b0:4005:808:0:0:0:200e"
	//query = "205.251.242.103"
	request, err := whois.Whois(query)
	if err == nil {
		fmt.Println(request)
		//result, err := whoisparser.Parse(request)
		re := regexp.MustCompile(`(?:Country:) (.*)`)
		country := re.FindAllStringSubmatch(string(request), -1)[0][1]
		fmt.Printf("%q\n", strings.TrimSpace(country))

		re_2 := regexp.MustCompile(`(?:OrgName:) (.*)`)
		org_name := re_2.FindAllStringSubmatch(string(request), -1)[0][1]
		fmt.Printf("%q\n", strings.TrimSpace(org_name))
		return
	} else {
		return "", "", err
	}

}