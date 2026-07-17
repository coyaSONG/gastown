package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/gastown/internal/beads"
)

func TestOutputMoleculeStatus_StandaloneFormulaShowsVars(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir tempDir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	status := MoleculeStatusInfo{
		HasWork:         true,
		PinnedBead:      &beads.Issue{ID: "gt-wisp-xyz", Title: "Standalone formula work"},
		AttachedFormula: "mol-release",
		AttachedVars:    []string{"version=1.2.3", "channel=stable"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputMoleculeStatus(status)

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	output := buf.String()

	if !strings.Contains(output, "📐 Formula: mol-release") {
		t.Fatalf("expected formula in output, got:\n%s", output)
	}
	if !strings.Contains(output, "--var version=1.2.3") || !strings.Contains(output, "--var channel=stable") {
		t.Fatalf("expected formula vars in output, got:\n%s", output)
	}
}

func TestOutputMoleculeStatus_FormulaWispShowsWorkflowContext(t *testing.T) {
	status := MoleculeStatusInfo{
		HasWork:         true,
		PinnedBead:      &beads.Issue{ID: "tool-wisp-demo", Title: "demo-hello"},
		AttachedFormula: "demo-hello",
		Progress: &MoleculeProgressInfo{
			RootID:     "tool-wisp-demo",
			RootTitle:  "demo-hello",
			TotalSteps: 3,
			DoneSteps:  0,
			ReadySteps: []string{"tool-wisp-step-1"},
		},
		NextAction: "Show the workflow steps: gt prime or bd mol current tool-wisp-demo",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputMoleculeStatus(status)

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	output := buf.String()

	if !strings.Contains(output, "📐 Formula: demo-hello") {
		t.Fatalf("expected formula line in output, got:\n%s", output)
	}
	if strings.Contains(output, "No molecule attached") {
		t.Fatalf("formula wisp should not be rendered as naked work, got:\n%s", output)
	}
	if strings.Contains(output, "Attach a molecule to start work") {
		t.Fatalf("formula wisp should not suggest gt mol attach, got:\n%s", output)
	}
	if !strings.Contains(output, "Show the workflow steps: gt prime or bd mol current tool-wisp-demo") {
		t.Fatalf("expected workflow next action, got:\n%s", output)
	}
}

func TestPopulateHookedWorkMetadata_RecognizesPatrolRoot(t *testing.T) {
	status := MoleculeStatusInfo{}
	bead := &beads.Issue{
		ID:          "gt-wisp-demo",
		Title:       "mol-witness-patrol (wisp)",
		Description: "Rendered patrol checklist without attachment metadata",
	}

	populateHookedWorkMetadata(&status, bead)

	if status.AttachedFormula != "mol-witness-patrol" {
		t.Fatalf("AttachedFormula = %q, want mol-witness-patrol", status.AttachedFormula)
	}
	if status.AttachedMolecule != "" {
		t.Fatalf("AttachedMolecule = %q, want empty for root-only patrol", status.AttachedMolecule)
	}
	if !status.IsWisp {
		t.Fatal("expected inferred patrol root to be marked as a wisp")
	}
}

func TestPopulateHookedWorkMetadata_PreservesExplicitFormula(t *testing.T) {
	status := MoleculeStatusInfo{}
	bead := &beads.Issue{
		Title:       "mol-witness-patrol custom work",
		Description: "attached_formula: operator-custom-patrol",
	}

	populateHookedWorkMetadata(&status, bead)

	if status.AttachedFormula != "operator-custom-patrol" {
		t.Fatalf("AttachedFormula = %q, want explicit metadata to win", status.AttachedFormula)
	}
}

func TestPopulateHookedWorkMetadata_DoesNotOverrideExplicitMolecule(t *testing.T) {
	status := MoleculeStatusInfo{}
	bead := &beads.Issue{
		Title:       "mol-witness-patrol custom work",
		Description: "attached_molecule: gt-wisp-explicit",
	}

	populateHookedWorkMetadata(&status, bead)

	if status.AttachedMolecule != "gt-wisp-explicit" {
		t.Fatalf("AttachedMolecule = %q, want explicit attachment", status.AttachedMolecule)
	}
	if status.AttachedFormula != "" {
		t.Fatalf("AttachedFormula = %q, want no inferred formula with explicit molecule", status.AttachedFormula)
	}
}
