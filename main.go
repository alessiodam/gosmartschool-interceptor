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
	domain := flag.String("domain", "", "Specify the Smartschool domain (e.g., school.smartschool.be)")

	flag.Parse()

	displayWarning()

	if !*autoYes && !confirmContinuation() {
		fmt.Println("Exiting the program.")
		return
	}

	if err := createRequestDirectory("./gsscap-requests"); err != nil {
		log.Fatalf("Error creating request directory: %v", err)
	}

	if err := promptForDomain(domain); err != nil {
		log.Fatalf("Error reading domain input: %v", err)
	}

	if err := validateDomain(*domain); err != nil {
		log.Fatalf("Error: %v", err)
	}

	logFilePath, err := interceptor.StartChromeAndCapture(*domain)
	if err != nil {
		log.Fatalf("Error starting Chrome and capturing requests: %v", err)
	}

	displaySessionEndMessage(logFilePath)
}

func displayWarning() {
	fmt.Println("WARNING: This tool will capture and save every HTTP request and response from your browsing session.")
}

func confirmContinuation() bool {
	fmt.Print("Do you want to continue? (yes/no): ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatalf("Error reading user input: %v", err)
	}

	return strings.ToLower(response) == "yes"
}

func createRequestDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func promptForDomain(domain *string) error {
	if *domain == "" {
		fmt.Print("Enter your Smartschool domain (e.g., school.smartschool.be): ")
		_, err := fmt.Scanln(domain)
		return err
	}
	return nil
}

func validateDomain(domain string) error {
	if !strings.HasSuffix(domain, "smartschool.be") && !strings.HasSuffix(domain, "smartschool.nl") {
		return fmt.Errorf("invalid domain. Please enter a valid Smartschool domain")
	}
	return nil
}

func displaySessionEndMessage(logFilePath string) {
	fmt.Println("Session ended. Check the log file at:", logFilePath)
	fmt.Println("Send the log file to @alessiodam on Discord to support GoSmartSchool development!")
	fmt.Println("These log files contain everything you did on Smartschool during the session, including potentially sensitive information like emails, usernames, passwords, etc. Cookies are NOT logged.")
	fmt.Println("If you wish, you can delete the sensitive info yourself before sending it while I'm working on an automatic solution.")
	fmt.Println("Thanks in advance! Your help is very valuable!")

	fmt.Print("Press enter to exit...")
	fmt.Scanln()
}
