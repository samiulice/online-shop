package main

import (
	"fmt"
	"regexp"
)

// IsBangladeshiMobileNumber checks if the given string is a valid Bangladeshi mobile number
func IsBangladeshiMobileNumber(mobileNumber string) bool {
	// Define a regex pattern for Bangladeshi mobile numbers
	pattern := `^01[3-9]\d{8}$`
	matched, err := regexp.MatchString(pattern, mobileNumber)
	if err != nil {
		// Handle error if the regex is invalid (unlikely)
		fmt.Println("Error matching regex:", err)
		return false
	}
	return matched
}

func main() {
	// Test cases
	testNumbers := []string{
		"01712345678", // valid
		"01312345678", // valid
		"01112345678", // invalid
		"0191234567",  // invalid (too short)
		"018123456789",// invalid (too long)
		"01123456782",// invalid (too long)
		"abcdefg",     // invalid (not a number)
	}

	for _, number := range testNumbers {
		if IsBangladeshiMobileNumber(number) {
			fmt.Printf("%s is a valid Bangladeshi mobile number.\n", number)
		} else {
			fmt.Printf("%s is NOT a valid Bangladeshi mobile number.\n", number)
		}
	}
}
