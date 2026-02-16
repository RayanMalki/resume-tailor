package main

import (
  "fmt"
  "os"
  "sort"

  "resume-tailor/internal/scoring/bm25"
  "resume-tailor/internal/scoring/classifier"
  "resume-tailor/internal/scoring/profiles"
)

func main() {
  resumeBytes, _ := os.ReadFile("/tmp/rayan_resume.txt")
  jobBytes, _ := os.ReadFile("/tmp/cn_job.txt")
  resume := string(resumeBytes)
  job := string(jobBytes)

  det := classifier.Detect(resume, job)
  prof := profiles.Get(det.Discipline)
  sig, err := bm25.ComputeWithProfile(resume, job, prof)
  if err != nil { panic(err) }

  fmt.Println("discipline:", det.Discipline, "confidence:", det.Confidence, "low:", det.LowConfidence)
  fmt.Printf("score: %.3f\n", sig.Score)
  fmt.Println("top missing terms:")
  for i, t := range sig.MissingJobTerms {
    if i >= 30 { break }
    fmt.Printf("- %s (%.2f) bucket=%s\n", t.Term, t.Score, t.Category)
  }

  fmt.Println("\nbucket coverage:")
  keys := make([]string, 0, len(sig.CategoryCoverage))
  for k := range sig.CategoryCoverage { keys = append(keys, k) }
  sort.Strings(keys)
  for _, k := range keys {
    fmt.Printf("- %s: %.2f\n", k, sig.CategoryCoverage[k])
  }

  fmt.Println("\nlow-signal filtered terms count:", len(sig.LowSignalTerms))
}
