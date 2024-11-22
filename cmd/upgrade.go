package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"io"
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
				return fmt.Errorf("opening new file: %w", err)
			}
			return downloadFile(link, f, ctx)

		} else {
			newBinary, err := util.TempFile("new-harvest-binary", 0777)
			if err != nil {
				return fmt.Errorf("cannot create new file: %w", err)
			}

			err = downloadFile(link, newBinary, ctx)
			if err != nil {
				return fmt.Errorf("downloading new file: %w", err)
			}

			orig, err := filepath.Abs(os.Args[0])
			if err != nil {
				return fmt.Errorf("something is wrong: %w", err)
			}

			execute, script, err := writeScript(orig, newBinary.Name())
			if err != nil {
				return fmt.Errorf("writing script: %w", err)
			}
			c := exec.Command(execute, script)
			return c.Start()
		}
	}),
}

func downloadFile(link string, destination *os.File, ctx context.Context) error {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return fmt.Errorf("creating latest version request: %w", err)
	}
	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("getting latest version: %w", err)
	}

	buff := bytes.Buffer{}
	size, err := io.Copy(&buff, res.Body)
	if err != nil {
		return fmt.Errorf("copying to buffer: %w", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), size)
	if err != nil {
		return fmt.Errorf("extracting zip: %w", err)
	}

	if len(reader.File) != 1 {
		return errors.New("unexpected number of files in zip")
	}

	newBinaryData, err := reader.File[0].Open()
	if err != nil {
		return fmt.Errorf("can't open downloaded file: %w", err)
	}

	if _, err := io.Copy(destination, newBinaryData); err != nil {
		return fmt.Errorf("can't write to new binary: %w", err)
	}

	return destination.Close()
}

func writeScript(oldBinary, newBinary string) (execute, script string, err error) {
	f, err := os.CreateTemp(os.TempDir(), "harvest-go-cli-upgrade")
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
		return "", fmt.Errorf("creating latest version request: %w", err)
	}
	req = req.WithContext(ctx)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("getting latest version: %w", err)
	}

	root, err := html.Parse(res.Body)
	if err != nil {
		return "", fmt.Errorf("parsing latest version response: %w", err)
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
