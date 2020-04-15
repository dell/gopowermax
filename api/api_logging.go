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
package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

func isBinOctetBody(h http.Header) bool {
	return h.Get(HeaderKeyContentType) == headerValContentTypeBinaryOctetStream
}

func logRequest(
	ctx context.Context,
	req *http.Request,
	lf func(func(args ...interface{}), string)) {

	w := &bytes.Buffer{}

	fmt.Fprintln(w)
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "POWERMAX HTTP REQUEST")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpRequest(req, !isBinOctetBody(req.Header))
	if err != nil {
		return
	}

	WriteIndented(w, buf)
	fmt.Fprintln(w)

	lf(log.Debug, w.String())
}

func logResponse(
	ctx context.Context,
	res *http.Response,
	lf func(func(args ...interface{}), string)) {

	w := &bytes.Buffer{}

	fmt.Fprintln(w)
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "POWERMAX HTTP RESPONSE")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpResponse(res, !isBinOctetBody(res.Header))
	if err != nil {
		return
	}

	bw := &bytes.Buffer{}
	WriteIndented(bw, buf)

	scanner := bufio.NewScanner(bw)
	for {
		if !scanner.Scan() {
			break
		}
		fmt.Fprintln(w, scanner.Text())
	}

	log.Debug(w.String())
}

// WriteIndentedN indents all lines n spaces.
func WriteIndentedN(w io.Writer, b []byte, n int) error {
	s := bufio.NewScanner(bytes.NewReader(b))
	if !s.Scan() {
		return nil
	}
	l := s.Text()
	for {
		for x := 0; x < n; x++ {
			if _, err := fmt.Fprint(w, " "); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, l); err != nil {
			return err
		}
		if !s.Scan() {
			break
		}
		l = s.Text()
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// WriteIndented indents all lines four spaces.
func WriteIndented(w io.Writer, b []byte) error {
	return WriteIndentedN(w, b, 4)
}
