package bm25

import "testing"

func TestComputeEmptyInputs(t *testing.T) {
	got, err := Compute("", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Score != 0 {
		t.Fatalf("expected score 0, got %v", got.Score)
	}
	if len(got.TopJobTerms) != 0 || len(got.MissingJobTerms) != 0 || len(got.OverlapTerms) != 0 {
		t.Fatalf("expected empty signals, got %+v", got)
	}
}

func TestComputeNoMissingTerms(t *testing.T) {
	resume := "Go developer with docker kubernetes"
	job := "Go developer docker kubernetes"

	got, err := Compute(resume, job)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.MissingJobTerms) != 0 {
		t.Fatalf("expected no missing terms, got %+v", got.MissingJobTerms)
	}
}

func TestComputeMissingTerms(t *testing.T) {
	resume := "Go developer"
	job := "Go developer kubernetes"

	got, err := Compute(resume, job)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.MissingJobTerms) == 0 {
		t.Fatalf("expected missing terms, got none")
	}
	if got.MissingJobTerms[0].Term != "kubernetes" {
		t.Fatalf("expected missing term kubernetes, got %+v", got.MissingJobTerms)
	}
}

func TestComputeDeterministicOrdering(t *testing.T) {
	resume := "alpha beta"
	job := "beta alpha"

	got, err := Compute(resume, job)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.TopJobTerms) < 2 {
		t.Fatalf("expected at least 2 top terms, got %+v", got.TopJobTerms)
	}

	first := got.TopJobTerms[0].Term
	second := got.TopJobTerms[1].Term
	if first != "alpha" || second != "beta" {
		t.Fatalf("expected deterministic alpha/beta order, got %s/%s", first, second)
	}
}
