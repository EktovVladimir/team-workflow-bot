package models

import "testing"

func TestIssueRef_ToUrl_DefaultOwner(t *testing.T) {
	ref := &IssueRef{Owner: "", Project: "PROJ", Number: "PROJ-123"}
	got := ref.ToUrl()
	want := "https://aviasales.atlassian.net/browse/PROJ-123"
	if got != want {
		t.Fatalf("ToUrl mismatch: got %s, want %s", got, want)
	}
}

func TestParseIssueRefFromUrl_Valid(t *testing.T) {
	cases := []struct {
		in      string
		owner   string
		project string
		number  string // полный ключ
	}{
		{"https://aviasales.atlassian.net/browse/PROJ-123", "aviasales", "PROJ", "PROJ-123"},
		{"http://aviasales.atlassian.net/browse/PROJ-456", "aviasales", "PROJ", "PROJ-456"},
		{"https://aviasales.atlassian.net/browse/PROJ-789/files", "aviasales", "PROJ", "PROJ-789"},
		{"https://aviasales.atlassian.net/browse/PROJ-1011?x=y#z", "aviasales", "PROJ", "PROJ-1011"},
	}
	for _, c := range cases {
		ref, ok := ParseIssueRefFromUrl(c.in)
		if !ok || ref == nil {
			t.Fatalf("expected ok for %q", c.in)
		}
		if ref.Owner != c.owner || ref.Project != c.project || ref.Number != c.number {
			t.Fatalf("parsed mismatch for %q: got %+v", c.in, ref)
		}
	}
}

func TestParseIssueRefFromUrl_Invalid(t *testing.T) {
	cases := []string{
		"https://example.atlassian.net/projects/PROJ/issues/123", // wrong path
		"https://aviasales.atlassian.net/browse/PROJ-",           // missing number
		"https://aviasales.atlassian.net/browse/-123",            // missing project
		"https://notjira.com/browse/PROJ-123",                    // wrong host
	}
	for _, in := range cases {
		ref, ok := ParseIssueRefFromUrl(in)
		if ok || ref != nil {
			t.Fatalf("expected fail for %q, got %+v", in, ref)
		}
	}
}
