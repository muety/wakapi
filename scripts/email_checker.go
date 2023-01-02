package main

// Usage example:
// cat emails.txt go run email_checker.go > result.txt

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

const MailPattern = "[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+"

var mailRegex *regexp.Regexp

func init() {
	mailRegex = regexp.MustCompile(MailPattern)
}

func CheckEmailMX(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	records, err := net.LookupMX(parts[1])
	return len(records) > 0 && err == nil
}

func ValidateEmail(email string) bool {
	return mailRegex.Match([]byte(email)) && CheckEmailMX(email)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		email := scanner.Text()
		if email == "" {
			return
		}

		if ValidateEmail(email) {
			fmt.Printf("[+] %s\n", email)
		} else {
			fmt.Printf("[-] %s\n", email)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
