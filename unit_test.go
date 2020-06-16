/*
 Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package pmax

import (
	"bufio"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/dell/gopowermax/mock"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

var mockServer *httptest.Server

func TestMain(m *testing.M) {
	status := 0
	var fileWriter *bufio.Writer
	var outputFile os.File
	format := "pretty"
	filename := ""
	testPaths := []string{"unittest"}
	testTags := ""
	runOptions := godog.Options{}
	regex, _ := regexp.Compile("(\\S+)=(\\S+)")

	// Read command line arguments
	for _, it := range os.Args[1:] {
		if allString := regex.FindAllStringSubmatch(it, 1); allString != nil {
			name := allString[0][1]
			value := allString[0][2]
			switch name {
			case "format":
				format = value
			case "outfile":
				filename = value
			case "test-paths":
				testPaths = strings.Split(value, ",")
			case "test-tags":
				testTags = value
			}
		}
	}

	// Create appropriate file name extension based on 'format' value
	if filename != "" {
		switch format {
		case "junit":
			filename = fmt.Sprintf("%s.xml", filename)
		case "cucumber":
			filename = fmt.Sprintf("%s.json", filename)
		case "pretty":
			filename = fmt.Sprintf("%s.txt", filename)
		case "progress":
			filename = fmt.Sprintf("%s.txt", filename)
		}

		outputFile, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not create output file %s - %v\n", filename, err)
			os.Exit(1)
		}
		fileWriter = bufio.NewWriter(outputFile)
		runOptions.Output = fileWriter
	}

	// Finalize the options
	runOptions.Format = format
	runOptions.Paths = testPaths
	runOptions.Tags = testTags

	// Start the mock server.
	handler := mock.GetHandler()
	mockServer = httptest.NewServer(handler)
	fmt.Printf("mockServer listening on %s\n", mockServer.URL)

	status = godog.RunWithOptions("godog", func(s *godog.Suite) {
		UnitTestContext(s)
	}, runOptions)

	if st := m.Run(); st > status {
		status = st
	}
	fmt.Printf("status %d\n", status)

	if fileWriter != nil {
		fileWriter.Flush()
	}
	outputFile.Close()
	os.Exit(status)
}
