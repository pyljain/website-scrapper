package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jung-kurt/gofpdf"
)

type Page struct {
	Title    string
	Content  string
	URL      string
	Headings []string
	Code     []string
}

func main() {
	// Define command-line flags
	baseURLFlag := flag.String("url", "", "The starting URL to scrape (required)")
	maxDepth := flag.Int("depth", 2, "Maximum depth for crawling links (default: 2)")
	outputFile := flag.String("output", "output.pdf", "Output PDF file name (default: output.pdf)")
	flag.Parse()

	// Validate URL
	if *baseURLFlag == "" {
		log.Fatal("Please provide a URL using the -url flag")
	}

	// Parse the URL to get the domain
	parsedURL, err := url.Parse(*baseURLFlag)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	// Extract the domain from the URL
	domain := parsedURL.Hostname()
	baseURL := *baseURLFlag
	pages := []Page{}
	visitedURLs := make(map[string]bool)

	// Initialize the collector
	c := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.MaxDepth(*maxDepth),
	)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("PDF Scraper", false)
	pdf.SetTitle("Go Blog Content", false)
	pdf.SetCreator("PDF Scraper", false)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n", r.URL.String())
	})

	// On every page
	c.OnHTML("div.Article, article", func(e *colly.HTMLElement) {
		currentURL := e.Request.URL.String()
		if visitedURLs[currentURL] {
			return
		}

		// Try different title selectors
		title := strings.TrimSpace(e.ChildText(".Header h1, h1"))
		if title == "" {
			title = strings.TrimSpace(e.ChildText(".Header h2, h2"))
		}
		if title == "" {
			title = "Untitled Article"
		}

		var content strings.Builder
		var headings []string
		var codeBlocks []string

		// Extract headings
		e.ForEach("h2, h3", func(_ int, el *colly.HTMLElement) {
			headings = append(headings, el.Text)
		})

		// Extract content with better formatting
		e.ForEach("p, pre, h2, h3, ul, ol", func(_ int, el *colly.HTMLElement) {
			switch el.Name {
			case "h2", "h3":
				content.WriteString("\n" + el.Text + "\n\n")
			case "p":
				content.WriteString(el.Text + "\n\n")
			case "pre":
				codeBlock := el.Text
				codeBlocks = append(codeBlocks, codeBlock)
				content.WriteString("[Code Block " + fmt.Sprintf("%d", len(codeBlocks)) + "]\n\n")
			case "ul", "ol":
				el.ForEach("li", func(_ int, li *colly.HTMLElement) {
					content.WriteString("• " + li.Text + "\n")
				})
				content.WriteString("\n")
			}
		})

		pages = append(pages, Page{
			Title:    title,
			Content:  content.String(),
			URL:      currentURL,
			Headings: headings,
			Code:     codeBlocks,
		})

		visitedURLs[currentURL] = true

		// Find and visit other links
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Attr("href")
			// Handle both absolute and relative URLs
			if strings.HasPrefix(link, "/") {
				// Relative URL
				absoluteURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, domain, link)
				if !visitedURLs[absoluteURL] {
					_ = c.Visit(absoluteURL)
				}
			} else if strings.HasPrefix(link, "http") {
				// Absolute URL - check if it's the same domain
				linkURL, parseErr := url.Parse(link)
				if parseErr == nil && linkURL.Hostname() == domain && !visitedURLs[link] {
					_ = c.Visit(link)
				}
			}
		})
	})

	// Start scraping
	err = c.Visit(baseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for all scraping to complete
	c.Wait()

	// Generate PDF with TOC
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(0, 10, "Table of Contents")
	pdf.Ln(20)

	// Create detailed TOC
	pdf.SetFont("Arial", "", 12)
	for i, page := range pages {
		// Main chapter entry
		pdf.SetFont("Arial", "B", 12)
		chapterNum := i + 1
		pdf.Cell(0, 10, fmt.Sprintf("%d. %s", chapterNum, page.Title))
		pdf.Ln(10)

		// Sub-sections
		pdf.SetFont("Arial", "", 10)
		for j, heading := range page.Headings {
			pdf.SetX(20) // Indent subsections
			pdf.Cell(0, 8, fmt.Sprintf("%d.%d. %s", chapterNum, j+1, heading))
			pdf.Ln(8)
		}
		pdf.Ln(5)
	}

	// Add content pages
	for i, page := range pages {
		pdf.AddPage()
		
		// Chapter title
		pdf.SetFont("Arial", "B", 20)
		pdf.Cell(0, 10, fmt.Sprintf("%d. %s", i+1, page.Title))
		pdf.Ln(15)

		// URL reference
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 10, "Source: "+page.URL)
		pdf.Ln(15)

		// Content
		pdf.SetFont("Arial", "", 12)
		
		// Split content into paragraphs and process each
		paragraphs := strings.Split(page.Content, "\n\n")
		for _, para := range paragraphs {
			if strings.TrimSpace(para) == "" {
				continue
			}
			
			// Check if it's a code block reference
			if strings.HasPrefix(para, "[Code Block ") {
				blockNum := 0
				fmt.Sscanf(para, "[Code Block %d]", &blockNum)
				if blockNum > 0 && blockNum <= len(page.Code) {
					// Add code block with monospace font and gray background
					pdf.SetFont("Courier", "", 10)
					pdf.SetFillColor(240, 240, 240)
					pdf.MultiCell(0, 5, page.Code[blockNum-1], "", "", true)
					pdf.SetFont("Arial", "", 12)
					pdf.SetFillColor(255, 255, 255)
					pdf.Ln(5)
				}
			} else {
				// Regular paragraph
				pdf.MultiCell(0, 6, para, "", "", false)
				pdf.Ln(3)
			}
		}
	}

	// Save the PDF
	// Ensure the output file has .pdf extension
	if !strings.HasSuffix(*outputFile, ".pdf") {
		*outputFile += ".pdf"
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputFile)
	if dirErr := os.MkdirAll(outputDir, 0755); dirErr != nil {
		log.Fatalf("Failed to create output directory: %v", dirErr)
	}

	err = pdf.OutputFileAndClose(*outputFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("PDF generated successfully with %d pages!\n", len(pages))
}
