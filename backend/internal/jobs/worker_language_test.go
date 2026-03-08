package jobs

import "testing"

func TestChooseCoverLetterLanguageSharedFrench(t *testing.T) {
	resumeText := "Le projet est realise avec des services cloud et des pipelines. Les responsabilites incluent le developpement et la gestion."
	jobText := "Le poste demande de l'experience avec des projets cloud, des API et des outils de monitoring."

	lang := chooseCoverLetterLanguage("French", resumeText, jobText)
	if lang != "French" {
		t.Fatalf("expected French, got %q", lang)
	}
}

func TestChooseCoverLetterLanguageFallsBackToJobWhenResumeUnknown(t *testing.T) {
	resumeText := "Go Docker Kubernetes."
	jobText := "The role requires experience with cloud platforms and building APIs in production systems."

	lang := chooseCoverLetterLanguage("", resumeText, jobText)
	if lang != "English" {
		t.Fatalf("expected English, got %q", lang)
	}
}

func TestChooseCoverLetterLanguageKeepsResumeLanguageOnConflict(t *testing.T) {
	resumeText := "Le projet est realise avec des services cloud et des pipelines. Les responsabilites incluent le developpement et la gestion."
	jobText := "The role requires experience with cloud platforms and building APIs in production systems."

	lang := chooseCoverLetterLanguage("French", resumeText, jobText)
	if lang != "French" {
		t.Fatalf("expected French, got %q", lang)
	}
}
