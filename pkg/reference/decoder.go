package reference

import (
	"errors"
	"regexp"
	"strings"
)

type (
	// ImageReference is parsed image reference specification
	ImageReference struct {
		Repository  string
		Tag         string
		RegistryURL string
	}
)

var (
	// ErrReferenceInvalidFormat is thrown when unable to parse image reference
	ErrReferenceInvalidFormat = errors.New("Invalid repository format")

	// Set of RegExp to decode Docker image reference
	tagRe    = regexp.MustCompile(`^:([a-z0-9._-]+)`)
	domainRe = regexp.MustCompile(`^(([a-z0-9-_]+)(\.[a-z0-9-_]+)*(:[0-9]+)?)`)
	pathRe   = regexp.MustCompile(`^/?(([a-z0-9-_.]+)(/[a-z0-9-_.]+)*)`)
)

// DecodeReference will try to parse image reference and return following structure:
//
// for given input:
// registry.example.com:80/sample/repository:42
//
// it will will ImageReference as follow:
// Repository: sample/repository
// Tag: 42
// RegistryURL: registry.example.com:80
//
// more examples in tests
func DecodeReference(reference string) (*ImageReference, error) {

	registryURL := ""
	if strings.Count(reference, ":") > 1 || strings.Count(reference, "/") > 0 {
		registryURL = domainRe.FindString(reference)
		reference = strings.Replace(reference, registryURL, "", 1)
	}

	repository := pathRe.FindString(reference)
	reference = strings.Replace(reference, repository, "", 1)
	repository = strings.TrimLeft(repository, "/")
	if repository == "" {
		return nil, ErrReferenceInvalidFormat
	}

	tag := tagRe.FindString(reference)
	tag = strings.TrimLeft(tag, ":")
	if tag == "" {
		return nil, ErrReferenceInvalidFormat
	}

	repo := &ImageReference{
		Repository:  repository,
		Tag:         tag,
		RegistryURL: registryURL,
	}

	return repo, nil
}
