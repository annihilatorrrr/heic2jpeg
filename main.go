package main

import (
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jdeng/goheif"
)

func convertHEICToJPEG(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := goheif.Decode(file)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return err
	}

	return nil
}

func processDirectory(inputDir, outputDir string) error {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file.Name())) == ".heic" {
			wg.Add(1)
			go func(file os.FileInfo) {
				defer wg.Done()
				inputPath := filepath.Join(inputDir, file.Name())
				outputPath := filepath.Join(outputDir, strings.TrimSuffix(file.Name(), ".heic")+".jpg")
				if err := convertHEICToJPEG(inputPath, outputPath); err != nil {
					errChan <- fmt.Errorf("failed to convert %s: %v", file.Name(), err)
				}
			}(file)
		}
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: heic_to_jpeg <input_directory> <output_directory>")
		return
	}

	inputDir := os.Args[1]
	outputDir := os.Args[2]

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	if err := processDirectory(inputDir, outputDir); err != nil {
		fmt.Printf("Error processing directory: %v\n", err)
		return
	}

	fmt.Println("Conversion completed successfully.")
}

