package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	gore "github.com/goretk/gore"
	"github.com/spf13/cobra"
)

var goVerIncompatible = "incompatible"
var noVersionStr = "v0.0.0"
var githubPrefix = "github.com"

type VendorInfo struct {
	PkgName string
	Date    time.Time
}

const longTmHelp = ` Calculate Approximate (minimum) timestamp

	There is a possibility to calculate timestamp using version of 
	3rd party dependency. 
	The algorithm:
	1. Get a list of dependencies
	2. Get the dependency version
	3. Get the date of a specific version (release) of the dependency
	4. Create a list of dates
	5. Take the latest date, it will be the approximate (minimum) timestamp

	More info in blogpost:
	https://www.orderofsixangles.com/en/2022/07/09/goelf-time-en.html
	https://www.orderofsixangles.com/ru/2022/07/09/goelf-time-ru.html
`

func init() {
	tmCmd := &cobra.Command{
		Use:     "tm path/to/go/file",
		Aliases: []string{"tm", "t"},
		Short:   "Calculate Approxmiate Timestamp.",
		Long:    longTmHelp,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			calcTm(args[0])
		},
	}

	rootCmd.AddCommand(tmCmd)
}

func calcTm(fileStr string) {

	fp, err := filepath.Abs(fileStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse the filepath: %s.\n", err)
		os.Exit(1)
	}

	f, err := gore.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when opening the file: %s.\n", err)
		os.Exit(1)
	}
	defer f.Close()

	pkgs, err := f.GetVendors()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when parsing packages: %s.\n", err)
		os.Exit(1)
	}

	vsInfo, err := GetVendorsInfo(pkgs)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get vendor info: %s.\n", err)
		os.Exit(1)
	}

	max := time.Time{}

	for _, vInfo := range vsInfo {
		fmt.Printf("%s %s\n", vInfo.PkgName, vInfo.Date.String())

		if vInfo.Date.After(max) {
			max = vInfo.Date
		}
	}

	fmt.Printf("\nApproximate (minimum) Timestamp equals = %s\n", max.String())

}

func getTagDate(owner string, repo string, client *github.Client, sha string) (time.Time, error) {

	commit, resp, err := client.Repositories.GetCommit(context.Background(), owner, repo, sha, nil)

	if err != nil {
		return time.Time{}, err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Failed to fetch %s/%s | status code = %d\n", owner, repo, resp.StatusCode)
		return time.Time{}, errors.New("failed to get commit")
	}

	return *commit.Commit.Committer.Date, nil
}

func ResolveReleaseDate(pName string, ver string) (time.Time, error) {

	//https://go.dev/blog/v2-go-modules
	// v2 postfix issue
	owner := strings.Split(pName, "/")[1]
	repo := strings.Split(pName, "/")[2]

	//fmt.Printf("fetching... owner = %s | repo = %s\n", owner, repo)

	// if u need more rate limit read https://github.com/google/go-github
	client := github.NewClient(nil)

	tags, resp, err := client.Repositories.ListTags(context.Background(), owner, repo, nil)

	if err != nil {
		return time.Time{}, err
	}

	if resp.StatusCode != 200 || len(tags) == 0 {
		fmt.Printf("Failed to fetch %s/%s | status code = %d\n", owner, repo, resp.StatusCode)
		return time.Time{}, err
	} else {
		for _, tag := range tags {
			if *tag.Name == ver {
				//fmt.Printf("Found tag %s\n", *tag.Name)

				return getTagDate(owner, repo, client, *tag.Commit.SHA)
			}
		}
	}

	return time.Time{}, nil
}

// strings like v0.0.0-20131221200532-179d4d0c4d8d
// mean that repo doesnt have releases. We should
// take 20131221 as a date
func parseNoVerStr(ver string) (time.Time, error) {

	var releaseDateStr string = (strings.Split(ver, "-")[1])[0:8]
	if len(releaseDateStr) != 0 {
		layout := "20060102"
		t, err := time.Parse(layout, releaseDateStr)

		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
	return time.Time{}, errors.New("сouldnt get date")
}

func getVersion(filePath string) string {
	i := strings.LastIndex(filePath, "@")
	if i != -1 && i != len(filePath)+1 {
		return strings.Split(filePath[i+1:], "/")[0]
	}
	return ""
}

func GetVendorsInfo(pkgs []*gore.Package) ([]VendorInfo, error) {

	var vsInfo []VendorInfo

	for _, p := range pkgs {

		var vInfo VendorInfo

		ver := getVersion(p.Filepath)

		if len(ver) == 0 || ver == "" {
			return nil, errors.New("failed to get version")
		}

		if strings.HasPrefix(p.Name, githubPrefix) {

			if strings.Contains(ver, goVerIncompatible) {
				ver = strings.Split(ver, "+")[0]
			}

			if strings.Contains(ver, noVersionStr) {
				t, err := parseNoVerStr(ver)
				if err != nil {
					fmt.Println(err)
					continue
				}
				vInfo = VendorInfo{p.Name, t}
				vsInfo = append(vsInfo, vInfo)
				continue
			}

			// get date using GitHub API
			//fmt.Println("Start fetching dates using Github API...")
			t, err := ResolveReleaseDate(p.Name, ver)
			if err != nil {
				fmt.Println(err)
				continue
			}
			vInfo = VendorInfo{p.Name, t}
			vsInfo = append(vsInfo, vInfo)

		} else {
			t, err := parseNoVerStr(ver)
			if err != nil {
				fmt.Println(err)
				continue
			}
			vInfo = VendorInfo{p.Name, t}
			vsInfo = append(vsInfo, vInfo)
		}
	}
	return vsInfo, nil
}
