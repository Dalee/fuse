package reference

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDecodeImage(t *testing.T) {
	var repo *ImageReference
	var err error

	refTestCaseList := []struct {
		input      string
		err        error
		repository string
		tag        string
		registry   string
	}{
		{
			input: "test_com",
			err: ErrReferenceInvalidFormat,
		},
		{
			input: "very.long.domain.registry:8080/test.com/repo:tag",
			repository: "test.com/repo",
			tag: "tag",
			registry: "very.long.domain.registry:8080",
		},
		{
			input: "example.com:5000/sample/unknown-repo:latest",
			repository: "sample/unknown-repo",
			tag: "latest",
			registry: "example.com:5000",
		},
		{
			input: "test.com:tag",
			registry: "",
			repository: "test.com",
			tag: "tag",
		},
		{
			input: "test.com:5000",
			registry: "",
			repository: "test.com",
			tag: "5000",
		},
		{
			input: "test.com/repo:tag",
			repository: "repo",
			tag: "tag",
			registry: "test.com",
		},
		{
			input: "test:5000/repo",
			err: ErrReferenceInvalidFormat,
		},
		{
			input: "test:5000/repo:tag",
			repository: "repo",
			tag: "tag",
			registry: "test:5000",
		},
		{
			input: "test:5000/repo:v1.2.3",
			repository: "repo",
			tag: "v1.2.3",
			registry: "test:5000",
		},
		{
			input: "test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			err: ErrReferenceInvalidFormat,
		},
		{
			input: "test:5000/repo:tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			repository: "repo",
			tag: "tag",
			registry: "test:5000",
		},
		{
			input: ":justtag",
			err: ErrReferenceInvalidFormat,
		},
		{
			input: "@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			err: ErrReferenceInvalidFormat,
		},
	}

	for _, testCase := range refTestCaseList {
		repo, err = DecodeReference(testCase.input)
		if err != testCase.err {
			t.Fatal(testCase.input, err)
		}

		if err != nil && repo != nil {
			t.Fatal("Error is not nil and repo is not nil", err, repo)
		}

		if err == testCase.err {
			continue
		}

		assert.Equal(t, repo.Repository, testCase.repository)
		assert.Equal(t, repo.Tag, testCase.tag)
		assert.Equal(t, repo.RegistryURL, testCase.registry)
	}
}
