package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: upload <path-to-pdf>")
		os.Exit(1)
	}
	pdfPath := os.Args[1]

	file, err := os.Open(pdfPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(pdfPath))
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(part, file); err != nil {
		panic(err)
	}
	writer.Close()

	resp, err := http.Post("http://localhost:8080/api/v1/pdf/split", writer.FormDataContentType(), body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	out, err := os.Create("split.zip")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		panic(err)
	}

	fmt.Println("Saved split.zip")
}
