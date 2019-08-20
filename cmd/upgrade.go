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
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var upgradeDownloadOnly string

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade harvest cli",
	Long:  `Download and install the latest binary`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		link, err := getDownloadLink(ctx)
		if err != nil {
			return err
		}

		if upgradeDownloadOnly != "" {
			f, err := os.OpenFile(upgradeDownloadOnly, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				return errors.Wrap(err, "opening new file")
			}
			return downloadFile(link, f, ctx)

		} else {
			newBinary, err := util.TempFile("new-harvest-binary", 0777)
			if err != nil {
				return errors.Wrap(err, "cannot create new file")
			}

			err = downloadFile(link, newBinary, ctx)
			if err != nil {
				return errors.Wrap(err, "downloading new file")
			}

			orig, err := filepath.Abs(os.Args[0])
			if err != nil {
				return errors.Wrap(err, "something is wrong")
			}

			execute, script, err := writeScript(orig, newBinary.Name())
			if err != nil {
				return errors.Wrap(err, "writing script")
			}
			c := exec.Command(execute, script)
			return c.Start()
		}
	}),
}

func downloadFile(link string, destination *os.File, ctx context.Context) error {
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
		return errors.Wrap(err, "copying to buffer")
	}

	reader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), size)
	if err != nil {
		return errors.Wrap(err, "extracting zip")
	}

	if len(reader.File) != 1 {
		return errors.New("unexpected number of files in zip")
	}

	newBinaryData, err := reader.File[0].Open()
	if err != nil {
		return errors.Wrap(err, "can't open downloaded file")
	}

	if _, err := io.Copy(destination, newBinaryData); err != nil {
		return errors.Wrap(err, "can't write to new binary")
	}

	return destination.Close()
}

func writeScript(oldBinary, newBinary string) (execute, script string, err error) {
	f, err := ioutil.TempFile(os.TempDir(), "harvest-go-cli-upgrade")
	if err != nil {
		return
	}

	var commands []string
	if runtime.GOOS == "windows" {
		return "", "", errors.New("windows upgrade not yet supported. Use --download-only")

	} else {
		execute = "sh"
		commands = []string{
			"#!/bin/bash",
			fmt.Sprintf("wait %d", os.Getpid()),
			fmt.Sprintf("cp %s %s", newBinary, oldBinary),
			fmt.Sprintf("rm %s", newBinary),
			`rm -- "$0"`,
		}
	}

	if _, err = f.WriteString(strings.Join(commands, "\n")); err != nil {
		return
	}
	script = f.Name()
	err = f.Close()
	return

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
	upgradeCmd.Flags().StringVar(&upgradeDownloadOnly, "download-only", "",
		"Just download the new binary in the current directory with the given name")
}
