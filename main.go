package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MaestroError/go-libheif"

	"github.com/julienschmidt/httprouter"
)

const maxFileSize = 500 << 20 // 500MB

func convertHEIC(inputPath, outputPath, format string) error {
	var err error
	switch format {
	case "jpeg":
		err = libheif.HeifToJpeg(inputPath, outputPath, 100)
	case "png":
		err = libheif.HeifToPng(inputPath, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return err
	}
	return nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}
	format := r.FormValue("format")
	if format != "jpeg" && format != "png" {
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("heic-convert-%d", time.Now().UnixNano()))
	if err = os.MkdirAll(tempDir, os.ModePerm); err != nil {
		log.Println(err)
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)
	errChan := make(chan error, len(files))
	var wg sync.WaitGroup
	if len(files) == 1 {
		fileHeader := files[0]
		wg.Add(1)
		go func(fileHeader *multipart.FileHeader) {
			defer wg.Done()
			file, err := fileHeader.Open()
			if err != nil {
				log.Println(err)
				errChan <- fmt.Errorf("failed to open file %s: %v", fileHeader.Filename, err)
				return
			}
			defer file.Close()
			inputPath := filepath.Join(tempDir, fileHeader.Filename)
			outputPath := filepath.Join(tempDir, strings.TrimSuffix(fileHeader.Filename, ".heic")+"."+format)
			out, err := os.Create(inputPath)
			if err != nil {
				log.Println(err)
				errChan <- fmt.Errorf("failed to create temp file %s: %v", inputPath, err)
				return
			}
			defer out.Close()
			if _, err = io.Copy(out, file); err != nil {
				log.Println(err)
				errChan <- fmt.Errorf("failed to save uploaded file %s: %v", inputPath, err)
				return
			}
			if err = convertHEIC(inputPath, outputPath, format); err != nil {
				log.Println(err)
				errChan <- fmt.Errorf("failed to convert file %s: %v", inputPath, err)
				return
			}
		}(fileHeader)
	} else {
		for _, fileHeader := range files {
			wg.Add(1)
			go func(fileHeader *multipart.FileHeader) {
				defer wg.Done()
				file, err := fileHeader.Open()
				if err != nil {
					log.Println(err)
					errChan <- fmt.Errorf("failed to open file %s: %v", fileHeader.Filename, err)
					return
				}
				defer file.Close()
				inputPath := filepath.Join(tempDir, fileHeader.Filename)
				outputPath := filepath.Join(tempDir, strings.TrimSuffix(fileHeader.Filename, ".heic")+"."+format)
				out, err := os.Create(inputPath)
				if err != nil {
					log.Println(err)
					errChan <- fmt.Errorf("failed to create temp file %s: %v", inputPath, err)
					return
				}
				defer out.Close()
				if _, err = io.Copy(out, file); err != nil {
					log.Println(err)
					errChan <- fmt.Errorf("failed to save uploaded file %s: %v", inputPath, err)
					return
				}
				if err = convertHEIC(inputPath, outputPath, format); err != nil {
					log.Println(err)
					errChan <- fmt.Errorf("failed to convert file %s: %v", inputPath, err)
					return
				}
			}(fileHeader)
		}
	}
	wg.Wait()
	close(errChan)
	for err = range errChan {
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// If there's more than one file, create the zip archive
	if len(files) > 1 {
		zipPath := filepath.Join(tempDir, "converted_files.zip")
		if err = createZip(tempDir, zipPath, format); err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("Failed to create zip archive: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename=converted_files.zip")
		http.ServeFile(w, r, zipPath)
	} else {
		// If there's only one file, serve it directly without zipping
		fileHeader := files[0]
		inputPath := filepath.Join(tempDir, fileHeader.Filename)
		w.Header().Set("Content-Disposition", "attachment; filename="+fileHeader.Filename)
		http.ServeFile(w, r, inputPath)
	}
}

func createZip(sourceDir, zipPath, format string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "."+format) {
			filePath := filepath.Join(sourceDir, file.Name())
			if err = addFileToZip(zipWriter, filePath, file.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath, filename string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	zipFile, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(zipFile, file)
	if err != nil {
		return err
	}
	return nil
}

func thehome(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("I'm alive!\nWelcome to HEIC to Image (JPEG/ PNG) Converter!\n\nBy @annihilatorrrr"))
}

func main() {
	router := httprouter.New()
	router.GET("/", thehome)
	router.POST("/convert", uploadHandler)
	handler := http.Handler(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9097"
	}
	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 180 * time.Second,
		Handler:      handler,
	}
	fmt.Println("Started!")
	if err := server.ListenAndServe(); err != nil {
		log.Println(err.Error())
	}
	log.Println("Bye!")
}
