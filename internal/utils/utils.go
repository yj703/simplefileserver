package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var client HTTPClient

func init() {

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

}

func LoadTemplateFile(location string, m *map[string]interface{}) ([]byte, error) {

	tmpl, err := template.ParseFiles(location)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Template file %v failed to be parsed. error: %v", location, err)
	}

	var templatedFileData bytes.Buffer

	err = tmpl.Execute(&templatedFileData, m)

	if err != nil {
		return nil, fmt.Errorf("ERROR: Template file %v failed to be templated. %v", location, err)
	}

	return templatedFileData.Bytes(), nil
}

func LoadTemplateFileToWriter(location string, m *map[string]interface{}, w io.Writer) (int64, error) {

	tmpl, err := template.ParseFiles(location)
	if err != nil {
		return 0, fmt.Errorf("ERROR: Template file %v failed to be parsed. error: %v", location, err)
	}

	var templatedFileData bytes.Buffer

	err = tmpl.Execute(&templatedFileData, m)

	if err != nil {
		return 0, fmt.Errorf("ERROR: Template file %v failed to be templated. %v", location, err)
	}

	return templatedFileData.WriteTo(w)
}

func ApplyTemplate(inputstr string, m *map[string]interface{}) string {

	tmpl, err := template.New("test").Parse(inputstr)
	if err != nil {
		log.Printf("template parsing error: %v", err)
		return inputstr
	}

	var templatedFileData bytes.Buffer

	err = tmpl.Execute(&templatedFileData, m)
	if err != nil {
		log.Printf("template execute error: %v", err)
		return inputstr
	}

	return string(templatedFileData.Bytes())

}

func PostJsonToAPI(apiurl string, token string, content string) ([]byte, error) {

	req, err := http.NewRequest("POST", apiurl, bytes.NewBufferString(content))

	if token != "" && (strings.Index(apiurl, ":8808") > 0 || strings.Index(apiurl, ":8806") > 0 || strings.Index(apiurl, ":8809") > 0) {
		req.Header.Set("Authorization", "Bearer *:"+token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Add("Content-Type", "application/json")
	//log.Printf("header:%v\n", req.Header)
	response, err := client.Do(req)
	if err != nil {
		//log.Printf("api error: %v\n", err)
		return nil, err
	} else {
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		//log.Printf("api return %v:%v\n", response.StatusCode, string(data))
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(string(data), "Error") {
			return nil, errors.New(string(data))
		}
		return data, nil
	}

}

func GetfromAPI(apiurl string, token string) ([]byte, error) {

	req, err := http.NewRequest("GET", apiurl, nil)
	if token != "" && (strings.Index(apiurl, ":8808") > 0 || strings.Index(apiurl, ":8806") > 0 || strings.Index(apiurl, ":8809") > 0) {
		req.Header.Set("Authorization", "Bearer *:"+token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(req)
	if err != nil {
		//log.Printf("api error: %v\n", err)
		return nil, err
	} else {
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		//log.Printf("api return %v\n", response.StatusCode)
		//fmt.Printf("%s\n",string(data))
		if strings.HasPrefix(string(data), "Error") {
			return nil, errors.New(string(data))
		}
		return data, nil
	}

}

func IsFileOrDirectoryPresent(directoryName string) bool {
	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		return false
	}
	return true
}
