package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"

	staticTemplate "github.com/TheShiveshNetwork/dizz/site/template"
)

type SiteData struct {
	Title       string
	Description string
	Domain      string
	Version     string
	Content     string
	Path        string
	IsDocs      bool
}

func renderMarkdown(content string) (string, error) {
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	var buf strings.Builder
	err := md.Convert([]byte(content), &buf)
	if err != nil {
		return "", err
	}

	// Sanitize HTML content
	htmlContent := sanitizeHTML(buf.String())

	// Post-process to add anchor IDs for headings
	htmlContent = strings.ReplaceAll(htmlContent, `<h2>Quick Start</h2>`, `<h2 id="quick-start">Quick Start</h2>`)
	htmlContent = strings.ReplaceAll(htmlContent, `<h2>Commands</h2>`, `<h2 id="commands">Commands</h2>`)
	htmlContent = strings.ReplaceAll(htmlContent, `<h2>Symbol States</h2>`, `<h2 id="symbol-states">Symbol States</h2>`)
	htmlContent = strings.ReplaceAll(htmlContent, `<h2>Architecture Overview</h2>`, `<h2 id="architecture">Architecture Overview</h2>`)

	return htmlContent, nil
}

// sanitizeHTML removes potentially dangerous HTML elements and attributes
func sanitizeHTML(content string) string {
	// Remove dangerous tags with word boundaries to avoid false positives
	dangerousTags := []string{
		`<script[^>]*>.*?</script>`,
		`<iframe[^>]*>.*?</iframe>`,
		`<object[^>]*>.*?</object>`,
		`<embed[^>]*>.*?</embed>`,
		`<form[^>]*>.*?</form>`,
		`<input[^>]*>`,
		`<textarea[^>]*>.*?</textarea>`,
	}

	for _, tag := range dangerousTags {
		re := regexp.MustCompile(`(?is)` + tag)
		content = re.ReplaceAllString(content, "")
	}

	// Remove dangerous event handlers and protocols - more precise matching
	dangerousAttrs := []string{
		`\bonload\s*=\s*"[^"]*"`,
		`\bonerror\s*=\s*"[^"]*"`,
		`\bonmouseover\s*=\s*"[^"]*"`,
		`\bonfocus\s*=\s*"[^"]*"`,
		`\bonblur\s*=\s*"[^"]*"`,
		`\bonchange\s*=\s*"[^"]*"`,
		`\bonsubmit\s*=\s*"[^"]*"`,
		`\bhref\s*=\s*["']javascript:`,
		`\bsrc\s*=\s*["']javascript:`,
	}

	for _, attr := range dangerousAttrs {
		re := regexp.MustCompile(`(?is)` + attr)
		content = re.ReplaceAllString(content, "")
	}

	return content
}

func readMarkdownFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getSiteData() SiteData {
	domain := "dizz.shitworks.co"
	release, _ := getLatestRelease()
	return SiteData{
		Title:       "Dizz",
		Description: "Progress-aware Dev CLI Tool. Know what to work on next.",
		Domain:      domain,
		Version:     release.Version,
		Path:        "/",
		IsDocs:      false,
	}
}

func getDocsData() SiteData {
	domain := "dizz.shitworks.co"
	release, _ := getLatestRelease()
	return SiteData{
		Title:       "Documentation",
		Description: "Complete documentation for Dizz - Progress-aware Dev CLI Tool",
		Domain:      domain,
		Version:     release.Version,
		Path:        "/docs",
		IsDocs:      true,
	}
}

type ReleaseInfo struct {
	Version string `json:"version"`
	Tag     string `json:"tag"`
}

func getLatestRelease() (*ReleaseInfo, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %v", err)
	}

	tag := strings.TrimSpace(string(output))
	version := strings.TrimPrefix(tag, "v")

	return &ReleaseInfo{
		Version: version,
		Tag:     tag,
	}, nil
}

func generatePage(data SiteData, markdownFile string, outputFile string) error {
	// Read markdown content
	markdownContent, err := readMarkdownFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading markdown file: %v", err)
	}

	// Convert markdown to HTML
	htmlContent, err := renderMarkdown(markdownContent)
	if err != nil {
		return fmt.Errorf("error rendering markdown: %v", err)
	}

	data.Content = htmlContent

	baseTemplate := staticTemplate.GetBaseTemplate()
	tmpl, err := template.New("page").Parse(baseTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(result.String()), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// @ignore-unused
func main() {
	// Generate home page
	homeData := getSiteData()
	if err := generatePage(homeData, "pages/index.md", "public/index.html"); err != nil {
		fmt.Printf("Error generating home page: %v\n", err)
		os.Exit(1)
	}

	// Generate docs page
	docsData := getDocsData()
	if err := generatePage(docsData, "pages/docs.md", "public/docs/index.html"); err != nil {
		fmt.Printf("Error generating docs page: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated pages successfully\n")
}
