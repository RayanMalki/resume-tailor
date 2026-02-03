package latex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CompilePDF(ctx context.Context, tectonicBin string, latex string) ([]byte, error) {
	if tectonicBin == "" {
		tectonicBin = "tectonic"
	}

	tmpDir, err := os.MkdirTemp("", "resume-tex-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	texPath := filepath.Join(tmpDir, "resume.tex")
	if err := os.WriteFile(texPath, []byte(latex), 0o600); err != nil {
		return nil, fmt.Errorf("write tex: %w", err)
	}

	compileCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()

	cmd := exec.CommandContext(compileCtx, tectonicBin, "-X", "compile", "--outdir", tmpDir, texPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tectonic failed: %w: %s", err, string(out))
	}

	pdfPath := filepath.Join(tmpDir, "resume.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}

	return pdfBytes, nil
}
