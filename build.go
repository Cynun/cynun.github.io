package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	once   sync.Once
	logger *slog.Logger
)

func GetLogger() *slog.Logger {
	once.Do(func() {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug})
		logger = slog.New(handler)
	})
	return logger
}
func chkerr(err error) {
	if err != nil {
		panic(err)
	}
}

type Article struct {
	Title string
	Link  string
}

func cleanUp(dir string) error {
	return exec.Command("bash", "-c", "rm -rf '"+dir+"/*'").Run()
}

func main() {
	rawDir := "raw"
	outDir := "html"
	universalTemplate := "universal.template.html"
	indexTemplate := "index.template.html"

	var articles []Article

	GetLogger()
	cleanUp(outDir)

	chkerr(filepath.Walk(rawDir, func(path string, info os.FileInfo, err error) error {
		chkerr(err)
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(rawDir, path)
		outPath := filepath.Join(outDir, strings.TrimSuffix(relPath, ".md")+".html")

		os.MkdirAll(filepath.Dir(outPath), 0755)

		logger.Debug("processing " + path)
		if strings.HasSuffix(path, ".md") {
			cmd := exec.Command("pandoc",
				"--standalone",
				"--css=/style.css",
				"--template="+universalTemplate,
				"--toc",
				"-o", outPath,
				path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			chkerr(cmd.Run())

			title := strings.TrimSuffix(relPath, ".md")
			link := filepath.ToSlash(strings.TrimPrefix(outPath, outDir+"/"))
			articles = append(articles, Article{Title: title, Link: link})
		} else {
			chkerr(exec.Command("cp", path, filepath.Join(outDir, relPath)).Run())
		}
		return nil
	}))

	// generate index.md
	var catalog strings.Builder
	for _, arc := range articles {
		catalog.WriteString("[" + arc.Title + "](" + arc.Link + ")\n\n")
	}
	cmd := exec.Command("pandoc",
		"--standalone",
		"--css=/style.css",
		"--template="+indexTemplate,
		"--toc",
		"-o", outDir+"/index.html")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(catalog.String())

	chkerr(cmd.Run())

	logger.Info("Generated " + strconv.Itoa(len(articles)) + " articles and index.html.")
}
