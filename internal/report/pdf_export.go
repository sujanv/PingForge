package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// GeneratePostScriptReport creates a print-ready vector PostScript / PDF printable document
func GeneratePostScriptReport(title string, results []core.ProbeResult, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("%!PS-Adobe-3.0\n")
	sb.WriteString("%%Title: " + title + "\n")
	sb.WriteString("%%Creator: PingForge Enterprise Diagnostics\n")
	sb.WriteString(fmt.Sprintf("%%%%CreationDate: %s\n", time.Now().Format(time.RFC1123)))
	sb.WriteString("%%Pages: 1\n")
	sb.WriteString("%%BoundingBox: 0 0 612 792\n")
	sb.WriteString("%%EndComments\n\n")

	sb.WriteString("/Helvetica-Bold findfont 18 scalefont setfont\n")
	sb.WriteString("50 740 moveto\n")
	sb.WriteString(fmt.Sprintf("(%s) show\n", title))

	sb.WriteString("/Helvetica findfont 10 scalefont setfont\n")
	sb.WriteString("50 720 moveto\n")
	sb.WriteString(fmt.Sprintf("(Generated: %s | Total Probes: %d) show\n", time.Now().Format("2006-01-02 15:04:05"), len(results)))

	// Draw horizontal line
	sb.WriteString("0.5 setlinewidth\n")
	sb.WriteString("50 710 moveto 560 710 lineto stroke\n\n")

	y := 680
	sb.WriteString("/Helvetica-Bold findfont 10 scalefont setfont\n")
	sb.WriteString(fmt.Sprintf("50 %d moveto (Target) show\n", y))
	sb.WriteString(fmt.Sprintf("200 %d moveto (Protocol) show\n", y))
	sb.WriteString(fmt.Sprintf("300 %d moveto (Status) show\n", y))
	sb.WriteString(fmt.Sprintf("400 %d moveto (Latency) show\n", y))

	sb.WriteString("/Helvetica findfont 9 scalefont setfont\n")
	for i, r := range results {
		if i >= 30 {
			break
		}
		y -= 18
		sb.WriteString(fmt.Sprintf("50 %d moveto (%s) show\n", y, r.Target.Host))
		sb.WriteString(fmt.Sprintf("200 %d moveto (%s) show\n", y, r.Type))
		sb.WriteString(fmt.Sprintf("300 %d moveto (%s) show\n", y, r.Status))
		sb.WriteString(fmt.Sprintf("400 %d moveto (%.2f ms) show\n", y, r.LatencyMs))
	}

	sb.WriteString("\nshowpage\n%%EOF\n")

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}
