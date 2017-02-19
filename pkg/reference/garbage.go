package reference

import (
	"fmt"
	"github.com/Dalee/hitman/pkg/registry"
	"strings"
)

type (
	// interface to hitman/pkg/registry
	registryInterface interface {
		GetImageDigestList(repo string) (*registry.RepositoryDigestList, error)
	}

	// GarbageDetectItem holds information about repository, deployed tags and garbage digests
	GarbageDetectItem struct {
		Repository        string
		DeployedTagList   []string
		GarbageDigestList []string
		GarbageTagList    []string
	}

	// GarbageDetectInfo holds whole list of GarbageDetectItem
	GarbageDetectInfo struct {
		Items []*GarbageDetectItem
	}
)

// StringInSlice checks is given string present in slice of strings
func StringInSlice(s string, sl []string) bool {
	for _, item := range sl {
		if strings.Compare(item, s) == 0 {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicated strings in slice (in-place)
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// SliceHasItemsInSlice is checks that any string in src slice contains in dst slice
func SliceHasItemsInSlice(src, dst []string) bool {
	for _, left := range src {
		for _, right := range dst {
			if strings.Compare(left, right) == 0 {
				return true
			}
		}
	}
	return false
}

// DetectGarbage will detect garbage for a given set of deployed image references
func DetectGarbage(k8sImageList []string, api registryInterface, ignoreMissing bool) (*GarbageDetectInfo, error) {
	// remove duplicated entries
	RemoveDuplicates(&k8sImageList)

	// prepare k8s deployed list and deployed repository list (to keep order, map order is not defined)
	deployedImages := make(map[string][]string, 0)
	deployedImagesList := make([]string, 0)

	for _, imageRefSpec := range k8sImageList {
		u, err := DecodeReference(imageRefSpec)
		if err != nil {
			return nil, err
		}

		deployedImages[u.Repository] =
			append(deployedImages[u.Repository], u.Tag)

		// if repository is not registered in orderList, register it
		if StringInSlice(u.Repository, deployedImagesList) == false {
			deployedImagesList =
				append(deployedImagesList, u.Repository)
		}
	}

	// prepare registry registered list
	registryImages := make(map[string][]*registry.RepositoryDigest)
	for repositoryPath := range deployedImages {

		imageInfo, err := api.GetImageDigestList(repositoryPath)
		if err != nil {
			// FIXME: either image is missed in repository or call failed
			// FIXME: make it more clear, right now - threat it as missing image
			if ignoreMissing == false {
				return nil, fmt.Errorf("Unknown image: %s", repositoryPath)
			}
			continue
		}

		// part of code, which is way to allow simulate situation
		// when registry answers something, but this something is not the image requested
		// (actually this should never happen in real life)
		found := true
		for _, digest := range imageInfo.Children {
			if strings.Compare(digest.Path, repositoryPath) != 0 {
				found = false
				break
			}
		}

		if found {
			registryImages[repositoryPath] =
				append(registryImages[repositoryPath], imageInfo.Children...)
		}
	}

	// build garbage list
	detectInfo := new(GarbageDetectInfo)
	for _, repositoryPath := range deployedImagesList {
		deployedTagList := deployedImages[repositoryPath]
		detectItem := &GarbageDetectItem{
			Repository:        repositoryPath,
			DeployedTagList:   deployedTagList,
			GarbageDigestList: []string{},
		}

		detectInfo.Items = append(detectInfo.Items, detectItem)

		imageDigestList, ok := registryImages[repositoryPath]
		if !ok {
			if ignoreMissing == false {
				return nil, fmt.Errorf("Unknown image: %s", repositoryPath)
			}
			continue
		}

		for _, digest := range imageDigestList {
			if SliceHasItemsInSlice(deployedTagList, digest.TagList) == false {
				detectItem.GarbageDigestList =
					append(detectItem.GarbageDigestList, digest.Name)

				detectItem.GarbageTagList =
					append(detectItem.GarbageTagList, digest.TagList...)
			}
		}
	}

	return detectInfo, nil
}
