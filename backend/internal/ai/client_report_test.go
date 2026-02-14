package ai

import "testing"

func TestUnmarshalJSONObjectWithRecoveryValid(t *testing.T) {
	var out ReportResponse
	in := `{"ats_report":{"score":0.4,"notes":["ok"],"summary":"s","interview_questions":[]},"change_plan":{"changes":[]}}`
	if err := unmarshalJSONObjectWithRecovery(in, &out); err != nil {
		t.Fatalf("expected valid json parse, got error: %v", err)
	}
	if out.ATSReport.Summary != "s" {
		t.Fatalf("unexpected summary: %q", out.ATSReport.Summary)
	}
}

func TestUnmarshalJSONObjectWithRecoveryTruncatedBrace(t *testing.T) {
	var out ReportResponse
	in := `{"ats_report":{"score":0.4,"notes":["ok"],"summary":"s","interview_questions":[]},"change_plan":{"changes":[]}`
	if err := unmarshalJSONObjectWithRecovery(in, &out); err != nil {
		t.Fatalf("expected truncated json to recover, got error: %v", err)
	}
}

func TestUnmarshalJSONObjectWithRecoveryExtractsObject(t *testing.T) {
	var out ReportResponse
	in := `Here is your JSON: {"ats_report":{"score":0.4,"notes":[],"summary":"s","interview_questions":[]},"change_plan":{"changes":[]}}`
	if err := unmarshalJSONObjectWithRecovery(in, &out); err != nil {
		t.Fatalf("expected wrapped json to recover, got error: %v", err)
	}
}
