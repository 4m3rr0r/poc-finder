package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"sync"
)

const (
	RED   = "\033[31m"
	GREEN = "\033[32m"
	BLUE  = "\033[34m"
	WHITE = "\033[37m"
)

const banner = `
┌─┐┌─┐┌─┐ ┌─┐┬┌┐┌┌┬┐┌─┐┬─┐
├─┘│ ││───├┤ ││││ ││├┤ ├┬┘
┴  └─┘└─┘ └  ┴┘└┘─┴┘└─┘┴└─                                                                              
`

const version = "1.0.0"

func logo() {
	fmt.Println(RED + banner + WHITE)
}

type CVE struct {
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
}

func getCVE(year string, cveID string) (*CVE, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/nomi-sec/PoC-in-GitHub/master/%s/%s.json", year, cveID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVE not found / other problem: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cveList []CVE
	err = json.Unmarshal(body, &cveList)
	if err != nil {
		return nil, err
	}

	if len(cveList) == 0 {
		return nil, fmt.Errorf("CVE not found / other problem")
	}

	return &cveList[0], nil
}

func extractYearFromCVE(cveID string) (string, error) {
	re := regexp.MustCompile(`CVE-(\d{4})-\d{4,}`)
	match := re.FindStringSubmatch(cveID)
	if len(match) != 2 {
		return "", fmt.Errorf("Invalid CVE format")
	}
	return match[1], nil
}

func cloneRepo(repoURL string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("git", "clone", repoURL)
	err := cmd.Run()
	if err != nil {
		fmt.Println(RED + "[+] Failed to clone repository: " + err.Error() + WHITE)
		return
	}
	fmt.Println(GREEN + "[+] Downloading the repository..." + WHITE)
}

func customUsage() {
	fmt.Println(RED + banner + WHITE)
	fmt.Println("Usage of cve_checker:")
	fmt.Println("  -cve string")
	fmt.Println("        Enter CVE Number Ex: -cve 'CVE-XXXX-XXXX'")
	fmt.Println("  -d")
	fmt.Println("        Clone the GitHub CVE repository")
	fmt.Println("  -v")
	fmt.Println("        Display version information")
}

func main() {
	cveID := flag.String("cve", "", "Enter CVE Number Ex: -cve 'CVE-XXXX-XXXX'")
	clone := flag.Bool("d", false, "Clone the GitHub CVE repository")
	showVersion := flag.Bool("v", false, "Display version information")

	// Set the custom usage function
	flag.Usage = customUsage

	flag.Parse()

	if *showVersion {
		fmt.Printf("CVE Checker Version: %s\n", version)
		return
	}

	if *cveID == "" {
		flag.Usage()
		return
	}

	year, err := extractYearFromCVE(*cveID)
	if err != nil {
		fmt.Println(RED + err.Error() + WHITE)
		return
	}

	logo()

	cve, err := getCVE(year, *cveID)
	if err != nil {
		fmt.Println(RED + err.Error() + WHITE)
		return
	}

	fmt.Println(GREEN + "[+] Git URL : " + BLUE + cve.HTMLURL + WHITE)
	fmt.Println(GREEN + "[+] Description : " + BLUE + cve.Description + WHITE)

	if *clone {
		var wg sync.WaitGroup
		wg.Add(1)
		go cloneRepo(cve.HTMLURL, &wg)
		wg.Wait()
		// If the clone was initiated, print success message after the program has compiled
		fmt.Println(GREEN + "[+] Successfully downloaded the repository." + WHITE)
	}
}
