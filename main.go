package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chai2010/tiff"
	"github.com/jdeng/goheif"
	"golang.org/x/image/draw"
)

func convertHEIC(inputPath, outputPath, format string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, err := goheif.Decode(file)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Copy(rgba, image.Point{}, img, bounds, draw.Src, nil)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	switch format {
	case "jpeg":
		err = jpeg.Encode(outFile, rgba, &jpeg.Options{Quality: 100})
	case "png":
		err = png.Encode(outFile, rgba)
	case "tiff":
		err = tiff.Encode(outFile, rgba, nil)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return err
	}
	return nil
}

func processDirectory(inputDir, outputDir, format string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(entries))
	for _, entry := range entries {
		if strings.ToLower(filepath.Ext(entry.Name())) == ".heic" {
			wg.Add(1)
			go func(entry os.DirEntry) {
				defer wg.Done()
				inputPath := filepath.Join(inputDir, entry.Name())
				outputPath := filepath.Join(outputDir, strings.TrimSuffix(entry.Name(), ".heic")+"."+format)
				if err := convertHEIC(inputPath, outputPath, format); err != nil {
					errChan <- fmt.Errorf("failed to convert %s: %v", entry.Name(), err)
				}
			}(entry)
		}
	}
	wg.Wait()
	close(errChan)
	for err = range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: heic_to_image <input_directory> <output_directory> <format>")
		return
	}
	inputDir := os.Args[1]
	outputDir := os.Args[2]
	format := os.Args[3]
	if format != "jpeg" && format != "png" && format != "tiff" {
		fmt.Printf("Unsupported format: %s. Supported formats are jpeg, png, and tiff.\n", format)
		return
	}
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}
	if err := processDirectory(inputDir, outputDir, format); err != nil {
		fmt.Printf("Error processing directory: %v\n", err)
		return
	}
	fmt.Println("Conversion completed successfully.")
}
