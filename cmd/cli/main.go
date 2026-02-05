package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  cli split -i <input.pdf> [-outdir <dir>] [-zip <zipfile>]")
	fmt.Println("  cli remove-pages -i <input.pdf> -o <output.pdf> -pages <selection> [-mode remove|keep]")
	fmt.Println("Examples:")
	fmt.Println("  cli split -i test.pdf -outdir output")
	fmt.Println("  cli split -i test.pdf -zip split.zip")
	fmt.Println("  cli remove-pages -i test.pdf -o result.pdf -pages '2,5-8,11'")
	fmt.Println("  cli remove-pages -i test.pdf -o result.pdf -pages '1,3,5' -mode keep")
}

func zipFiles(zipPath string, files []string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	for _, f := range files {
		fr, err := os.Open(f)
		if err != nil {
			return err
		}
		_, name := filepath.Split(f)
		fw, err := zw.Create(name)
		if err != nil {
			fr.Close()
			return err
		}
		if _, err := io.Copy(fw, fr); err != nil {
			fr.Close()
			return err
		}
		fr.Close()
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "split":
		fs := flag.NewFlagSet("split", flag.ExitOnError)
		in := fs.String("i", "", "input PDF file")
		outdir := fs.String("outdir", "", "output directory to move parts to")
		zipPath := fs.String("zip", "", "optional zip file to write parts into")
		fs.Parse(os.Args[2:])

		if *in == "" {
			fmt.Println("input file required")
			fs.Usage()
			os.Exit(2)
		}

		parts, err := pdf.SplitPDFFile(*in)
		if err != nil {
			log.Fatalf("split failed: %v", err)
		}

		// If outdir specified, move files
		if *outdir != "" {
			if err := os.MkdirAll(*outdir, 0755); err != nil {
				log.Fatalf("failed create outdir: %v", err)
			}
			var moved []string
			for _, p := range parts {
				_, name := filepath.Split(p)
				dst := filepath.Join(*outdir, name)
				if err := os.Rename(p, dst); err != nil {
					log.Fatalf("failed move part: %v", err)
				}
				moved = append(moved, dst)
			}
			parts = moved
		}

		if *zipPath != "" {
			if err := zipFiles(*zipPath, parts); err != nil {
				log.Fatalf("failed create zip: %v", err)
			}
			fmt.Printf("wrote %s\n", *zipPath)
		} else {
			fmt.Println("parts:")
			for _, p := range parts {
				fmt.Println(p)
			}
		}

	case "remove-pages":
		fs := flag.NewFlagSet("remove-pages", flag.ExitOnError)
		in := fs.String("i", "", "input PDF file")
		out := fs.String("o", "", "output PDF file")
		pages := fs.String("pages", "", "page selection, e.g. '2,5-8,11'")
		mode := fs.String("mode", "remove", "mode: 'remove' (default) or 'keep'")
		fs.Parse(os.Args[2:])

		if *in == "" || *out == "" || *pages == "" {
			fmt.Println("input, output and pages are required")
			fs.Usage()
			os.Exit(2)
		}

		keepMode := *mode == "keep"
		result, err := pdf.RemovePagesFromFile(*in, *out, *pages, keepMode)
		if err != nil {
			log.Fatalf("remove-pages failed: %v", err)
		}

		fmt.Printf("Mode: %s\n", result["mode"])
		fmt.Printf("Original pages: %d\n", result["original_pages"])
		fmt.Printf("Removed: %d pages\n", result["removed_count"])
		fmt.Printf("Remaining: %d pages\n", result["remaining_pages"])
		fmt.Printf("Output: %s\n", result["output_path"])

	default:
		usage()
		os.Exit(1)
	}
}
