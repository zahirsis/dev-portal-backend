package unix

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	templateHtml "html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type directoryService struct {
	logger logger.Logger
}

func NewDirectoryService(logger logger.Logger) service.DirectoryService {
	return &directoryService{
		logger: logger,
	}
}

func (d *directoryService) CopyDirectory(src string, dest string) error {
	cmd := exec.Command("cp", "-r", src, dest)
	return d.execCommand(cmd, fmt.Sprintf("copying %s to %s", src, dest))
}

func (d *directoryService) CopyFile(src string, dest string) error {
	cmd := exec.Command("cp", src, dest)
	return d.execCommand(cmd, fmt.Sprintf("copying %s to %s", src, dest))
}

func (d *directoryService) CreateDirectory(path string) error {
	cmd := exec.Command("mkdir", "-p", path)
	return d.execCommand(cmd, fmt.Sprintf("creating directory %s", path))
}

func (d *directoryService) RemoveDirectory(path string) error {
	cmd := exec.Command("rm", "-rf", path)
	return d.execCommand(cmd, fmt.Sprintf("removing directory %s", path))
}

func (d *directoryService) DirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		d.logger.Error("Error checking directory", path, err.Error())
		return false, err
	}
}

func (d *directoryService) VerifyOrInsertLineInFile(path string, line string) error {
	inputFile, err := os.Open(path)
	if err != nil {
		d.logger.Error("Error opening file", path, err.Error())
		return err
	}
	var buf bytes.Buffer
	scanner := bufio.NewScanner(inputFile)
	exists := false
	for scanner.Scan() {
		l := scanner.Text()
		if l != "" {
			buf.WriteString(l + "\n")
		}
		if strings.Trim(l, " ") == strings.Trim(line, " ") {
			exists = true
		}
	}
	err = inputFile.Close()
	if err != nil {
		d.logger.Error("Error closing file", path, err.Error())
		return err
	}
	if !exists {
		buf.WriteString(line + "\n")
	}
	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		d.logger.Error("Error writing file", path, err.Error())
		return err
	}
	return nil
}

func (d *directoryService) ApplyTemplate(path string, values interface{}) (err error) {
	d.logger.Debug("Applying template on file", path)
	defer func() {
		if rec := recover(); rec != nil {
			if e, ok := rec.(error); ok {
				d.logger.Error("Error applying template", path, e.Error())
				err = e
				return
			}
			d.logger.Error("Unknown error applying template", path, rec)
		}
		return
	}()
	var buf bytes.Buffer
	t := template.Must(template.ParseFiles(path))
	err = t.Execute(&buf, values)
	if err != nil {
		d.logger.Error("Error creating file from template", path, err.Error())
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		d.logger.Error("Error writing file", path, err.Error())
		return err
	}
	return d.ClearBlankLinesFromFile(path)
}

func (d *directoryService) LoadTemplate(path string, values interface{}, html bool) (value []byte, err error) {
	d.logger.Debug("Loading template on file", path)
	defer func() {
		if rec := recover(); rec != nil {
			if e, ok := rec.(error); ok {
				d.logger.Error("Error loading template", path, e.Error())
				err = e
				return
			}
			d.logger.Error("Unknown error loading template", path, rec)
		}
		return
	}()
	var buf bytes.Buffer
	if html {
		t := templateHtml.Must(templateHtml.ParseFiles(path))
		err = t.Execute(&buf, values)
		if err != nil {
			d.logger.Error("Error loading result content from template", path, err.Error())
			return nil, err
		}
	} else {
		t := template.Must(template.ParseFiles(path))
		err = t.Execute(&buf, values)
		if err != nil {
			d.logger.Error("Error loading result content from template", path, err.Error())
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (d *directoryService) ApplyTemplateRecursively(rootDir string, values interface{}) error {
	return filepath.Walk(rootDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			d.logger.Error("Error walking file", filePath, err.Error())
			return err
		}
		if filePath == rootDir {
			if !fileInfo.IsDir() {
				err := d.ApplyTemplate(filePath, values)
				if err != nil {
					return err
				}
			}
			return nil
		}
		if fileInfo.IsDir() {
			return d.ApplyTemplateRecursively(filePath, values)
		}
		if err = d.ApplyTemplate(filePath, values); err != nil {
			return err
		}
		return nil
	})
}

func (d *directoryService) RenameFile(src string, dest string) error {
	cmd := exec.Command("mv", src, dest)
	return d.execCommand(cmd, fmt.Sprintf("renaming file %s to %s", src, dest))
}

func (d *directoryService) DeleteFile(path string) error {
	cmd := exec.Command("rm", "-f", path)
	return d.execCommand(cmd, fmt.Sprintf("deleting file %s", path))
}

func (d *directoryService) ClearBlankLinesFromFile(path string) error {
	inputFile, err := os.Open(path)
	if err != nil {
		d.logger.Error("Error opening file", path, err.Error())
		return err
	}
	var buf bytes.Buffer
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Trim(line, " ") != "" {
			buf.WriteString(line + "\n")
		}
	}
	err = inputFile.Close()
	if err != nil {
		d.logger.Error("Error closing file", path, err.Error())
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		d.logger.Error("Error writing file", path, err.Error())
		return err
	}
	return nil
}

func (d *directoryService) execCommand(cmd *exec.Cmd, action string) error {
	d.logger.Debug(action)
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		d.logger.Error(fmt.Sprintf("Error %s", action), err.Error())
	}
	scanner := bufio.NewScanner(stderr)
	var errMessage string
	for scanner.Scan() {
		errMessage += scanner.Text() + "\n"
	}
	if errMessage != "" {
		d.logger.Error(fmt.Sprintf("Error %s", action), errMessage)
		return errors.New(fmt.Sprintf("Error %s: %s", action, errMessage))
	}
	return nil
}
