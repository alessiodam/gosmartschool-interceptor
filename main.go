package main

import (
	"flag"
	"fmt"
	"gosmartschool-interceptor/interceptor"
	"log"
	"os"
	"strings"
)

func main() {
	autoYes := flag.Bool("y", false, "Automatically confirm continuation")
	domain := flag.String("domain", "", "Specify the smartschool domain (ex. school.smartschool.be)")

	flag.Parse()

	fmt.Println("WARNING: This tool will capture and save every HTTP request and response from your browsing session.")

	if !*autoYes {
		fmt.Print("Do you want to continue? (yes/no): ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			log.Fatalf("Error reading user input: %v", err)
		}

		if response != "yes" {
			fmt.Println("Exiting the program.")
			return
		}
	}

	if _, err := os.Stat("./gsscap-requests"); os.IsNotExist(err) {
		err = os.Mkdir("./gsscap-requests", 0755)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if *domain == "" {
		fmt.Print("Enter your smartschool domain (ex. school.smartschool.be): ")
		_, err := fmt.Scanln(domain)
		if err != nil {
			log.Fatalf("Error reading user input: %v", err)
		}
	}
	if !strings.HasSuffix(*domain, "smartschool.be") && !strings.HasSuffix(*domain, "smartschool.nl") {
		_, _ = fmt.Fprintln(os.Stderr, "Error: Invalid domain. Please enter a valid smartschool domain.")
		os.Exit(1)
	}
	err := interceptor.StartChromeAndCapture(*domain)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session ended. Check the './gsscap-requests' folder for the saved requests and responses.")
}
