# Web to PDF Scraper

A Go-based tool that scrapes web content and generates a well-formatted PDF with a table of contents, chapters, and code blocks. The tool is particularly useful for creating offline documentation or e-books from web content.

## Features

- Scrapes web content from a specified URL
- Generates a structured PDF with:
  - Table of Contents
  - Chapter-based organization
  - Sub-sections based on page headings
  - Code block formatting with monospace font and gray background
  - Source URL references
- Configurable crawling depth
- Support for both relative and absolute URLs
- Custom output file naming

## Installation

1. Ensure you have Go 1.21 or later installed
2. Clone this repository
3. Install dependencies:
```bash
go mod tidy
```

## Usage

Run the program with the following command-line flags:

```bash
go run main.go -url <starting-url> [-depth <depth>] [-output <output-file>]
```

### Command Line Options

- `-url` (required): The starting URL to scrape
- `-depth` (optional): Maximum depth for crawling links (default: 2)
- `-output` (optional): Output PDF file name (default: "output.pdf")

### Example

```bash
# Scrape a website with default settings
go run main.go -url https://example.com/docs

# Scrape with custom depth and output file
go run main.go -url https://example.com/docs -depth 3 -output documentation.pdf
```

## Dependencies

- [github.com/gocolly/colly/v2](https://github.com/gocolly/colly): Web scraping framework
- [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf): PDF generation library

## Output Format

The generated PDF includes:

1. **Cover Page**: Title and basic information
2. **Table of Contents**: List of all scraped pages with their sections
3. **Content Pages**: Each scraped page is formatted as a chapter with:
   - Chapter title
   - Source URL reference
   - Formatted content including:
     - Regular paragraphs
     - Section headings
     - Code blocks with special formatting
     - Bullet points and numbered lists

## Limitations

- Only scrapes content from the same domain as the starting URL
- Some complex JavaScript-rendered content may not be captured
- PDF formatting is optimized for article-style content

## Contributing

Feel free to open issues or submit pull requests for improvements or bug fixes.

## License

MIT License - feel free to use and modify as needed.
