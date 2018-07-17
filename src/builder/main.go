package main

import (
	"bytes"
	"encoding/json"
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
				stemcellVersion = fmt.Sprintf("%v", stemcell["version"])
				if stemcell["os"] == "ubuntu-trusty" && stemcellVersion != "<nil>" {
					if _, err := DownloadStemcell(stemcellVersion); err != nil {
						return "", err
					}
				}
			}
		}
	}
	return stemcellVersion, nil
}

func isStemcellUploaded(stemcellVersion string) (bool, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bosh", "stemcells", "--json")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
		return false, err
	}
	data := struct {
		Tables []struct {
			Rows []struct {
				Version string `json:"version"`
			}
		}
	}{}
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		return false, err
	}
	for _, table := range data.Tables {
		for _, row := range table.Rows {
			if row.Version == stemcellVersion || row.Version == fmt.Sprintf("%s*", stemcellVersion) {
				return true, nil
			}
		}
	}
	return false, nil
}

func UploadStemcell(stemcellVersion string) error {
	if uploaded, err := isStemcellUploaded(stemcellVersion); err != nil {
		return err
	} else if uploaded {
		return nil
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		"bosh", "upload-stemcell",
		fmt.Sprintf("https://s3.amazonaws.com/bosh-gce-light-stemcells/light-bosh-stemcell-%s-google-kvm-ubuntu-trusty-go_agent.tgz", stemcellVersion),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
		return err
	}
	return nil
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
				} else if url, ok := release["url"].(string); ok && url != "<nil>" {
					if strings.HasPrefix(url, "file://release") {
						release["url"] = "file:///var/vcap/cache" + (release["url"].(string))[6:]
						fmt.Println("Convert Absolute:", release["url"])
					} else if release["stemcell"] != nil || strings.Contains(url, "-compiled-") {
						fmt.Println("Download:", url)
						if err := Download(url, path); err != nil {
							return fmt.Errorf("download: %s: %s", url, err)
						}
						release["url"] = newURL
					} else {
						fmt.Println("Compile:", url)
						if err := CompileRelease(stemcellVersion, release, path); err != nil {
							return fmt.Errorf("compile release: %s: %s", url, err)
						}
						release["url"] = newURL
					}
				}
			}
		}
	}
	return nil
}

func OneInstance(data Yaml) {
	if groups, ok := data["instance_groups"].([]interface{}); ok {
		for _, group := range groups {
			if group, ok := group.(Yaml); ok {
				if i, ok := group["instances"].(int); ok {
					if i > 1 {
						group["instances"] = 1
						fmt.Printf("Instances: %s: %d -> 1\n", group["name"], i)
					} else {
						fmt.Printf("Instances: %s: %d\n", group["name"], i)
					}
				}
			}
		}
	}
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
		fmt.Printf("===\n=== expected stemcell %s, found %s (using found version)\n===\n", stemcellVersion, foundStemcellVersion)
		stemcellVersion = foundStemcellVersion
	}
	if err := UploadStemcell(stemcellVersion); err != nil {
		return fmt.Errorf("upload stemcell: %s: %s", file, err)
	}
	if err := Releases(data, stemcellVersion); err != nil {
		return fmt.Errorf("releases: %s: %s", file, err)
	}
	OneInstance(data)
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
