package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sans-sroc/integrity/pkg/common"
	"github.com/sans-sroc/integrity/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type createCommand struct {
}

func (w *createCommand) Execute(c *cli.Context) error {
	dir := c.String("directory")
	ver := c.String("courseware-version")
	jsonOut := c.Bool("json")
	pretty := c.Bool("json-pretty")
	user := c.String("user")

	files := []common.File{}

	versionFileName := fmt.Sprintf("VERSION-%s.txt", ver)
	versionPartFileName := fmt.Sprintf("VERSION-%s-part.txt", ver)
	versionFirstFileName := fmt.Sprintf("VERSION-%s-first.txt", ver)

	fileVersionPath := path.Join(dir, versionFileName)
	fileVersionPartPath := path.Join(dir, versionPartFileName)
	fileVersionFirstPath := path.Join(dir, versionFirstFileName)

	getFirstPath := path.Join(dir, "get_first")
	getFirstExists := false
	getFirstIsEmpty := true

	if !jsonOut {
		fmt.Println("[+] Working directory:", dir)

		if _, err := os.Stat(fileVersionPath); err == nil {
			fmt.Println("[!] VERSION file already exists! Overwriting!")
			err = os.Remove(fileVersionPath)
			if err != nil {
				logrus.WithError(err).Error("Cannot delete VERSION file")
				return err
			}
		}

		if _, err := os.Stat(fileVersionPartPath); err == nil {
			if err := os.Remove(fileVersionPartPath); err != nil {
				logrus.WithError(err).Error("Cannot delete VERSION file")
				return err
			}
		}

		if _, err := os.Stat(fileVersionFirstPath); err == nil {
			if err = os.Remove(fileVersionFirstPath); err != nil {
				logrus.WithError(err).Error("Cannot delete VERSION file")
				return err
			}
		}

		if _, err := os.Stat(getFirstPath); err == nil {
			isEmpty, err := utils.IsDirectoryEmpty(getFirstPath)
			if err != nil {
				return err
			}

			getFirstExists = true
			getFirstIsEmpty = isEmpty

			if !isEmpty {
				if err := utils.CreateVerFile(fileVersionPartPath, user); err != nil {
					return err
				}

				if err := utils.CreateVerFile(fileVersionFirstPath, user); err != nil {
					return err
				}
			}
		}

		if err := utils.CreateVerFile(fileVersionPath, user); err != nil {
			return err
		}
	}

	if err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			pathCheck, err := os.Stat(path)
			if err != nil {
				logrus.WithError(err).Error("Cannot process file")
				return err
			}

			if !pathCheck.IsDir() {
				hash, err := utils.HashFileSha256(path)
				if err != nil {
					logrus.WithError(err).Error("Cannot hash file")
					return err
				}

				fileName, err := filepath.Rel(dir, path)
				if err != nil {
					logrus.WithError(err).Error("Cannot determine file path")
					return err
				}

				files = append(files, common.File{
					Name: filepath.ToSlash(fileName),
					Hash: hash,
				})
			}

			return nil
		},
	); err != nil {
		logrus.WithError(err).Error("Validation failed")
		return err
	}

	if jsonOut {
		var out []byte
		var err error
		if pretty {
			out, err = json.MarshalIndent(common.FileOutput{Files: files}, "", "    ")
			if err != nil {
				logrus.WithError(err).Error("unable to render json")
				return err
			}
		} else {
			out, err = json.Marshal(common.FileOutput{Files: files})
			if err != nil {
				logrus.WithError(err).Error("unable to render json")
				return err
			}
		}

		fmt.Println(string(out))

		return nil
	} else {
		for _, file := range files {
			if !strings.Contains(file.Name, fmt.Sprintf("VERSION-%s", ver)) {
				fmt.Println("[+] Processing " + file.Name + "...")
				utils.AppendVerFile(fileVersionPath, fileVersionPartPath, fileVersionFirstPath, file.Name, file.Hash, dir, getFirstExists, getFirstIsEmpty)
			}
		}

	}

	return nil
}

func init() {
	cmd := createCommand{}

	cliCmd := &cli.Command{
		Name:   "create",
		Usage:  "create integrity files",
		Action: cmd.Execute,
		Flags:  globalFlags(),
		Before: globalBefore,
	}

	common.RegisterCommand(cliCmd)
}
