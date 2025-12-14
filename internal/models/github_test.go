package models

import "testing"

func TestParsePullRequestRefFromUrl_ValidVariants(t *testing.T) {
	cases := []struct {
		in    string
		owner string
		repo  string
		num   int
	}{
		{"https://github.com/org/repo/pull/123", "org", "repo", 123},
		{"http://github.com/org/repo/pull/456", "org", "repo", 456},
		{"https://github.com/org/repo/pull/789/", "org", "repo", 789},
		{"https://github.com/org/repo/pull/1011/files", "org", "repo", 1011},
		{"https://github.com/org/repo/pull/1213?some=param#anchor", "org", "repo", 1213},
	}
	for _, c := range cases {
		ref, ok := ParsePullRequestRefFromUrl(c.in)
		if !ok || ref == nil {
			t.Fatalf("expected ok for %q", c.in)
		}
		if ref.Owner != c.owner || ref.Repo != c.repo || ref.Number != c.num {
			t.Fatalf("parsed mismatch for %q: got %+v", c.in, ref)
		}
	}
}

func TestParsePullRequestRefFromUrl_Invalid(t *testing.T) {
	cases := []string{
		"https://github.com/org/repo/pulls/123",      // plural path
		"https://github.com/org/repo/issues/123",     // not pull
		"https://example.com/org/repo/pull/123",      // not github.com
		"https://github.com//repo/pull/1",            // missing owner
		"https://github.com/org//pull/1",             // missing repo
		"https://github.com/org/repo/pull/notnumber", // not numeric
		"github.com/org/repo/pull/",                  // missing number
	}
	for _, in := range cases {
		ref, ok := ParsePullRequestRefFromUrl(in)
		if ok || ref != nil {
			t.Fatalf("expected fail for %q, got %+v", in, ref)
		}
	}
}
