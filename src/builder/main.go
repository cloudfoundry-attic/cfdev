package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Yaml map[interface{}]interface{}

func Parse(file string) (Yaml, error) {
	txt, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %s", err)
	}
	data := make(Yaml)
	if err := yaml.Unmarshal(txt, &data); err != nil {
		return nil, fmt.Errorf("parse file: %s", err)
	}
	return data, err
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Download(url, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error while downloading %s - %s", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Non 200 status code: %d: %s", response.StatusCode, url)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir for file: %s", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open file for write: %s", err)
	}
	if _, err := io.Copy(f, response.Body); err != nil {
		return fmt.Errorf("copy data to file: %s", err)
	}
	return f.Close()
}

func DownloadStemcell(version string) (string, error) {
	url := fmt.Sprintf("https://s3.amazonaws.com/bosh-core-stemcells/warden/bosh-stemcell-%s-warden-boshlite-ubuntu-trusty-go_agent.tgz", version)
	path := filepath.Join("output/cache", fmt.Sprintf("bosh-stemcell-%s-warden-boshlite-ubuntu-trusty-go_agent.tgz", version))
	newURL := fmt.Sprintf("file:///var/vcap/cache/%s", filepath.Base(path))
	if Exists(path) {
		fmt.Println("Skip Stemcell:", version)
		return newURL, os.Chtimes(path, time.Now(), time.Now())
	}
	fmt.Println("Download Stemcell:", version)
	err := Download(url, path)
	return newURL, err
}

func Stemcells(data Yaml) (string, error) {
	stemcellVersion := ""
	if pools, ok := data["resource_pools"].([]interface{}); ok {
		for _, pool := range pools {
			if pool, ok := pool.(Yaml); ok {
				if stemcell, ok := pool["stemcell"].(Yaml); ok {
					m := regexp.MustCompile(`v=([\d\.]+)$`).FindStringSubmatch(stemcell["url"].(string))
					if len(m) == 2 {
						stemcellVersion = m[1]
					} else {
						return "", fmt.Errorf("couldn't find stemcell version: %s: %v", stemcell["url"], m)
					}
					var err error
					stemcell["url"], err = DownloadStemcell(stemcellVersion)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}
	if stemcells, ok := data["stemcells"].([]interface{}); ok {
		for _, stemcell := range stemcells {
			if stemcell, ok := stemcell.(Yaml); ok {
				if stemcell["os"] == "ubuntu-trusty" {
					stemcellVersion = stemcell["version"].(string)
					if _, err := DownloadStemcell(stemcellVersion); err != nil {
						return "", err
					}
				}
			}
		}
	}
	return stemcellVersion, nil
}

func CompileRelease(stemcellVersion string, release map[interface{}]interface{}, path string) error {
	// Upload Manifest to Bosh
	manifest, err := yaml.Marshal(map[string]interface{}{
		"instance_groups": []interface{}{},
		"name":            "cf",
		"releases":        []map[interface{}]interface{}{release},
		"stemcells": []interface{}{
			map[string]string{
				"alias":   "default",
				"os":      "ubuntu-trusty",
				"version": stemcellVersion,
			},
		},
		"update": map[string]interface{}{
			"canaries":          0,
			"canary_watch_time": "30000-1200000",
			"max_in_flight":     0,
			"update_watch_time": "5000-1200000",
		},
	})
	if err != nil {
		return fmt.Errorf("marhal yaml: %s", err)
	}
	cmd := exec.Command("bosh", "-n", "deploy", "-d", "cf", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	txt, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(txt))
		return fmt.Errorf("bosh deploy: %s", err)
	}

	// Download Release
	tmpDir, err := ioutil.TempDir("", "extract-releases-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	if txt, err := exec.Command(
		"bosh", "-d", "cf", "export-release",
		fmt.Sprintf("%v/%v", release["name"], release["version"]),
		fmt.Sprintf("ubuntu-trusty/%s", stemcellVersion),
		"--dir", tmpDir,
	).CombinedOutput(); err != nil {
		fmt.Println(string(txt))
		return err
	}
	m, err := filepath.Glob(filepath.Join(tmpDir, "*"))
	if err != nil {
		return err
	}
	if len(m) != 1 {
		return fmt.Errorf("Could not find single file: %v", m)
	}
	return exec.Command("mv", m[0], path).Run()
}

func Releases(data Yaml, stemcellVersion string) error {
	if releases, ok := data["releases"].([]interface{}); ok {
		for _, release := range releases {
			if release, ok := release.(Yaml); ok {
				path := fmt.Sprintf("output/cache/releases/%v-%v-%s.tgz", release["name"], release["version"], stemcellVersion)
				newURL := fmt.Sprintf("file:///var/vcap/cache/releases/%v-%v-%s.tgz", release["name"], release["version"], stemcellVersion)
				if Exists(path) {
					fmt.Println("Skip:", filepath.Base(path))
					if err := os.Chtimes(path, time.Now(), time.Now()); err != nil {
						return err
					}
					release["url"] = newURL
				} else if release["stemcell"] != nil || strings.Contains(release["url"].(string), "-compiled-") {
					fmt.Println("Download:", release["url"])
					if err := Download(release["url"].(string), path); err != nil {
						return fmt.Errorf("download: %s: %s", release["url"], err)
					}
					release["url"] = newURL
				} else {
					fmt.Println("Compile:", release["url"])
					if err := CompileRelease(stemcellVersion, release, path); err != nil {
						return fmt.Errorf("compile release: %s: %s", release["url"], err)
					}
					release["url"] = newURL
				}
			}
		}
	}
	return nil
}

func Write(data interface{}, path string) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func Process(stemcellVersion, file string) error {
	data, err := Parse(file)
	if err != nil {
		return fmt.Errorf("parse file: %s: %s", file, err)
	}
	foundStemcellVersion, err := Stemcells(data)
	if err != nil {
		return fmt.Errorf("stemcells: %s: %s", file, err)
	}
	if foundStemcellVersion != "" && foundStemcellVersion != stemcellVersion {
		// TODO have a better plan
		return fmt.Errorf("expected stemcell %s, found %s", stemcellVersion, foundStemcellVersion)
	}
	if err := Releases(data, stemcellVersion); err != nil {
		return fmt.Errorf("releases: %s: %s", file, err)
	}
	if err := Write(data, file); err != nil {
		return fmt.Errorf("write: %s: %s", file, err)
	}
	return nil
}

func main() {
	stemcellVersion := os.Args[1]
	for _, file := range os.Args[2:] {
		fmt.Println("===", file)
		if err := Process(stemcellVersion, file); err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}
}
