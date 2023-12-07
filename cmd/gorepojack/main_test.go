package main

import "testing"

func TestInferRepoURL(t *testing.T) {
	testcases := []struct {
		repo string
		want string
	}{
		{"github.com/user/repo", "https://github.com/user/repo"},
		{"", ""},
		{"github.com/DataDog/datadog-go/v5", "https://github.com/DataDog/datadog-go"},
		{"github.com/aws/aws-sdk-go-v2/service/kms", "https://github.com/aws/aws-sdk-go-v2"},
	}
	for _, tc := range testcases {
		got := inferRepoURL(tc.repo)
		if tc.want != got {
			t.Errorf("want %q, got %q", tc.want, got)
		}
	}
}

func TestInferUserURL(t *testing.T) {
	repo := "github.com/user/repo"
	want := "https://github.com/user"
	got := inferUserURL(repo)
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}
