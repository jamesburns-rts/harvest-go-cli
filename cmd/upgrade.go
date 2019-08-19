/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// settable values

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade harvest cli",
	Long:  `TODO - longer description`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		link, err := getDownloadLink(ctx)
		if err != nil {
			return err
		}
		return downloadFile(link, ctx)
	}),
}

func downloadFile(link string, ctx context.Context) error {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return errors.Wrap(err, "creating latest version request")
	}
	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "getting latest version")
	}

	buff := bytes.Buffer{}
	size, err := io.Copy(&buff, res.Body)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), size)
	if err != nil {
		return err
	}

	if len(reader.File) != 1 {
		return errors.New("unexpected number of files in zip")
	}

	newFileName := "new_harvest_cli"
	newBinary, err := os.OpenFile(newFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return errors.Wrap(err, "cannot create new file")
	}
	newBinaryData, err := reader.File[0].Open()
	if err != nil {
		return errors.Wrap(err, "can't open downloaded file")
	}
	if _, err := io.Copy(newBinary, newBinaryData); err != nil {
		return errors.Wrap(err, "can't write to current binary")
	}

	orig, err := filepath.Abs(os.Args[0])
	if err != nil {
		return errors.Wrap(err, "something is wrong")
	}

	newFileName, err = filepath.Abs(newFileName)
	if err != nil {
		return errors.Wrap(err, "something is wrong")
	}
	fmt.Printf("cp %s %s && rm %s\n", newFileName, orig, newFileName)

	//binary, err := os.OpenFile(os.Args[0], os.O_RDWR, 0)
	//if err != nil {
	//	return errors.Wrap(err, "can't open current binary")
	//}
	//newBinary, err := reader.File[0].Open()
	//if err != nil {
	//	return errors.Wrap(err, "can't open downloaded file")
	//}
	//if _, err := io.Copy(binary, newBinary); err != nil {
	//	return errors.Wrap(err, "can't write to current binary")
	//}

	// write to file
	// cp newFile oldFile
	// rm newFile

	return nil
}

func getDownloadLink(ctx context.Context) (string, error) {
	req, err := http.NewRequest("GET", "https://github.com/jamesburns-rts/harvest-go-cli/releases/latest", nil)
	if err != nil {
		return "", errors.Wrap(err, "creating latest version request")
	}
	req = req.WithContext(ctx)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "getting latest version")
	}

	root, err := html.Parse(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "parsing latest version response")
	}

	// parse projects
	node := htmlquery.FindOne(root, "/html/body/a")
	if node == nil || node.FirstChild == nil {
		return "", errors.New("parsing latest version response - something has changed")
	}

	for _, attr := range node.Attr {
		if attr.Key == "href" {
			path := strings.ReplaceAll(attr.Val, "/tag/", "/download/")

			// get file name
			goos := runtime.GOOS
			if goos == "darwin" {
				goos = "mac"
			}
			return fmt.Sprintf("%s/%s_%s.zip", path, goos, runtime.GOARCH), nil

		}
	}

	return "", errors.New("no link href")
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
