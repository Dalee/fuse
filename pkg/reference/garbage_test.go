package reference

import (
	"errors"
	"github.com/Dalee/hitman/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type (
	RegistryInterfaceMock struct {
		mock.Mock
	}
)

func (rm *RegistryInterfaceMock) GetImageDigestList(repo string) (*registry.RepositoryDigestList, error) {
	var digestList *registry.RepositoryDigestList

	args := rm.Called(repo)
	passedList := args.Get(0)
	if passedList != nil {
		digestList = passedList.(*registry.RepositoryDigestList)
	}

	return digestList, args.Error(1)
}

func TestSliceHasItemsInSlice(t *testing.T) {
	assert.True(t, SliceHasItemsInSlice([]string{"1", "2"}, []string{"2", "3"}))
	assert.False(t, SliceHasItemsInSlice([]string{"1", "2"}, []string{"3", "4"}))
}

func TestRemoveDuplicates(t *testing.T) {
	dups := &[]string{"1", "1", "2", "2"}
	RemoveDuplicates(dups)

	assert.Equal(t, []string{"1", "2"}, *dups)
}

func TestStringInSlice(t *testing.T) {
	assert.True(t, StringInSlice("hello", []string{"world", "hello"}))
	assert.False(t, StringInSlice("example", []string{"world", "hello"}))
}

//
func TestDetectGarbage_NormalCase(t *testing.T) {

	//
	// setup registry answers for sample/repo1
	//
	registryList1 := new(registry.RepositoryDigestList)
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-1",
		Path:    "sample/repo1",
		TagList: []string{"3"},
	})
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-2",
		Path:    "sample/repo1",
		TagList: []string{"4"},
	})
	// this one is wrong and should be deleted
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-3",
		Path:    "sample/repo1",
		TagList: []string{"5", "latest"},
	})

	//
	// setup registry answers for sample/repo2
	//
	registryList2 := new(registry.RepositoryDigestList)
	registryList2.Children = append(registryList2.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo2-randomdigestnumber-1",
		Path:    "sample/repo2",
		TagList: []string{"v26"},
	})
	registryList2.Children = append(registryList2.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo2-randomdigestnumber-2",
		Path:    "sample/repo2",
		TagList: []string{"v27"},
	})
	// this one is wrong
	registryList2.Children = append(registryList2.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo2-randomdigestnumber-3",
		Path:    "sample/repo2",
		TagList: []string{"v28", "latest"},
	})

	// build mock
	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/repo1").Return(registryList1, nil)
	registryMock.On("GetImageDigestList", "sample/repo2").Return(registryList2, nil)

	// k8s deployments
	deployedList := []string{
		"example.com:5000/sample/repo1:3",
		"example.com:5000/sample/repo1:4",
		"example.com:5000/sample/repo2:v26",
		"example.com:5000/sample/repo2:v27",
	}

	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, false)
	assert.Nil(t, err)

	assert.Len(t, garbageInfo.Items, 2)

	garbageItem1 := garbageInfo.Items[0]
	assert.Equal(t, "sample/repo1", garbageItem1.Repository)
	assert.Equal(t, []string{"sha256:sample-repo1-randomdigestnumber-3"}, garbageItem1.GarbageDigestList)
	assert.Equal(t, []string{"3", "4"}, garbageItem1.DeployedTagList)
	assert.Equal(t, []string{"5", "latest"}, garbageItem1.GarbageTagList)

	garbageItem2 := garbageInfo.Items[1]
	assert.Equal(t, "sample/repo2", garbageItem2.Repository)
	assert.Equal(t, []string{"sha256:sample-repo2-randomdigestnumber-3"}, garbageItem2.GarbageDigestList)
	assert.Equal(t, []string{"v26", "v27"}, garbageItem2.DeployedTagList)
	assert.Equal(t, []string{"v28", "latest"}, garbageItem2.GarbageTagList)
}

func TestDetectGarbage_SkipTags(t *testing.T) {
	registryList1 := new(registry.RepositoryDigestList)
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-1",
		Path:    "sample/repo1",
		TagList: []string{"3"},
	})
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-2",
		Path:    "sample/repo1",
		TagList: []string{"4"},
	})
	// this one is wrong and should be deleted
	registryList1.Children = append(registryList1.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo1-randomdigestnumber-3",
		Path:    "sample/repo1",
		TagList: []string{"5", "latest"},
	})

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/repo1").Return(registryList1, nil)

	deployedList := []string{
		"example.com:5000/sample/repo1:3",
		"example.com:5000/sample/repo1:4",
	}

	// check
	garbageInfo, err := DetectGarbage(deployedList, []string{"latest"}, registryMock, false)
	assert.Nil(t, err)
	assert.Len(t, garbageInfo.Items, 1)

	garbageItem := garbageInfo.Items[0]
	assert.Equal(t, []string{"3", "4"}, garbageItem.DeployedTagList)
	assert.Equal(t, []string(nil), garbageItem.GarbageTagList)
}

//
func TestDetectGarbage_RegistryCallFailed(t *testing.T) {
	deployedList := []string{
		"example.com:5000/sample/repo:latest",
	}

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/repo").Return(nil, errors.New("Call failed"))

	//
	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, false)
	assert.Error(t, err)
	assert.Nil(t, garbageInfo)
}

//
func TestDetectGarbage_RegistryCallFailedNoError(t *testing.T) {
	deployedList := []string{
		"example.com:5000/sample/repo:latest",
	}

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/repo").Return(nil, errors.New("Call failed"))

	//
	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, true)
	assert.Nil(t, err)
	assert.Len(t, garbageInfo.Items, 1)

	garbageItem := garbageInfo.Items[0]
	assert.Equal(t, "sample/repo", garbageItem.Repository)
	assert.Equal(t, []string{"latest"}, garbageItem.DeployedTagList)
	assert.Equal(t, []string{}, garbageItem.GarbageDigestList)
}

//
func TestDetectGarbage_ImageNotFoundError(t *testing.T) {
	deployedList := []string{
		"example.com:5000/sample/unknown-repo:latest",
	}

	registryList := new(registry.RepositoryDigestList)
	registryList.Children = append(registryList.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo2-randomdigestnumber-1",
		Path:    "sample/repo",
		TagList: []string{"latest", "v1.2.3"},
	})

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/unknown-repo").Return(registryList, nil)

	//
	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, false)
	assert.Error(t, err)
	assert.Nil(t, garbageInfo)
}

//
func TestDetectGarbage_ImageNotFoundSkipMissing(t *testing.T) {
	deployedList := []string{
		"example.com:5000/sample/unknown-repo:latest",
	}

	registryList := new(registry.RepositoryDigestList)
	registryList.Children = append(registryList.Children, &registry.RepositoryDigest{
		Name:    "sha256:sample-repo2-randomdigestnumber-1",
		Path:    "sample/repo",
		TagList: []string{"latest", "v1.2.3"},
	})

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/unknown-repo").Return(registryList, nil)

	//
	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, true)
	assert.Nil(t, err)
	assert.Len(t, garbageInfo.Items, 1)

	garbageItem := garbageInfo.Items[0]
	assert.Empty(t, garbageItem.GarbageDigestList)
	assert.Equal(t, "sample/unknown-repo", garbageItem.Repository)
	assert.Equal(t, []string{"latest"}, garbageItem.DeployedTagList)
}

func TestDetectGarbage_InvalidRefSpecPassed(t *testing.T) {
	deployedList := []string{
		"example.com/sample/unknown-repo",
	}

	registryMock := new(RegistryInterfaceMock)
	registryMock.On("GetImageDigestList", "sample/unknown-repo").Return(nil, errors.New("Shouldn't be there"))

	garbageInfo, err := DetectGarbage(deployedList, []string{}, registryMock, true)
	assert.Nil(t, garbageInfo)
	assert.Error(t, err)
	assert.Equal(t, "Invalid repository format", err.Error())
}
