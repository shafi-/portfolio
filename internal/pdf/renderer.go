package pdf

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
)

// Renderer creates PDF documents
type Renderer struct {
	pdf *gofpdf.Fpdf
}

// NewRenderer creates a new PDF renderer
func NewRenderer() *Renderer {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetMargins(20, 15, 20)

	return &Renderer{pdf: pdf}
}

// RenderCV renders a CV to PDF and returns the file path
func (r *Renderer) RenderCV(markdown string, outputPath string) error {
	r.pdf.AddPage()

	// Parse and render markdown sections
	lines := strings.Split(markdown, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			i++
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "## ") {
			sectionTitle := strings.TrimPrefix(line, "## ")
			r.renderSectionHeader(sectionTitle)
			i++
			continue
		}

		// Bold text (name, company, etc.)
		if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			text := strings.Trim(line, "*")
			r.renderBoldText(text)
			i++
			continue
		}

		// Italic text (dates, location)
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
			text := strings.Trim(line, "*")
			r.renderItalicText(text)
			i++
			continue
		}

		// Bullet points
		if strings.HasPrefix(line, "- ") {
			text := strings.TrimPrefix(line, "- ")
			r.renderBullet(text)
			i++
			continue
		}

		// Mixed formatting (e.g., "**Key:** Value")
		if strings.Contains(line, "**") {
			r.renderFormattedLine(line)
			i++
			continue
		}

		// Regular text
		r.renderText(line)
		i++
	}

	return r.pdf.OutputFileAndClose(outputPath)
}

func (r *Renderer) renderSectionHeader(title string) {
	r.pdf.Ln(4)
	r.pdf.SetFont("Helvetica", "B", 14)
	r.pdf.SetTextColor(33, 37, 41)
	r.pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	r.pdf.SetDrawColor(33, 37, 41)
	r.pdf.Line(20, r.pdf.GetY(), 190, r.pdf.GetY())
	r.pdf.Ln(2)
}

func (r *Renderer) renderBoldText(text string) {
	r.pdf.SetFont("Helvetica", "B", 11)
	r.pdf.SetTextColor(33, 37, 41)
	r.pdf.CellFormat(0, 6, text, "", 1, "L", false, 0, "")
}

func (r *Renderer) renderItalicText(text string) {
	r.pdf.SetFont("Helvetica", "I", 10)
	r.pdf.SetTextColor(108, 117, 125)
	r.pdf.CellFormat(0, 5, text, "", 1, "L", false, 0, "")
}

func (r *Renderer) renderBullet(text string) {
	r.pdf.SetFont("Helvetica", "", 10)
	r.pdf.SetTextColor(33, 37, 41)
	x := r.pdf.GetX()
	r.pdf.CellFormat(5, 5, "•", "", 0, "L", false, 0, "")
	r.pdf.SetX(x + 5)
	r.pdf.MultiCell(165, 5, text, "", "L", false)
}

func (r *Renderer) renderFormattedLine(line string) {
	r.pdf.SetFont("Helvetica", "", 10)
	r.pdf.SetTextColor(33, 37, 41)

	// Parse bold sections
	parts := strings.Split(line, "**")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i%2 == 1 {
			// Bold part
			r.pdf.SetFont("Helvetica", "B", 10)
			r.pdf.CellFormat(0, 5, part, "", 0, "L", false, 0, "")
		} else {
			// Normal part
			r.pdf.SetFont("Helvetica", "", 10)
			r.pdf.CellFormat(0, 5, part, "", 0, "L", false, 0, "")
		}
	}
	r.pdf.Ln(5)
}

func (r *Renderer) renderText(text string) {
	r.pdf.SetFont("Helvetica", "", 10)
	r.pdf.SetTextColor(33, 37, 41)
	r.pdf.MultiCell(0, 5, text, "", "L", false)
}

// RenderCVToBytes renders a CV to PDF bytes
func (r *Renderer) RenderCVToBytes(markdown string) ([]byte, error) {
	r.pdf.AddPage()

	lines := strings.Split(markdown, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			i++
			continue
		}

		if strings.HasPrefix(line, "## ") {
			sectionTitle := strings.TrimPrefix(line, "## ")
			r.renderSectionHeader(sectionTitle)
			i++
			continue
		}

		if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			text := strings.Trim(line, "*")
			r.renderBoldText(text)
			i++
			continue
		}

		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
			text := strings.Trim(line, "*")
			r.renderItalicText(text)
			i++
			continue
		}

		if strings.HasPrefix(line, "- ") {
			text := strings.TrimPrefix(line, "- ")
			r.renderBullet(text)
			i++
			continue
		}

		if strings.Contains(line, "**") {
			r.renderFormattedLine(line)
			i++
			continue
		}

		r.renderText(line)
		i++
	}

	var buf strings.Builder
	err := r.pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to render PDF: %w", err)
	}

	return []byte(buf.String()), nil
}
