package models

import (
	"io"
)

type Resource struct {
	PluginID int    `json:"id"`
	Scheme   string `json:"scheme"`
	Body     string `json:"body"`
}

type Metadata struct {
	Name      string `yaml:"name,omitempty"`       // 参数为尽力提供，可能不存在
	Size      int64  `yaml:"size,omitempty"`       // 参数为尽力提供，可能不存在
	MediaType string `json:"media_type,omitempty"` // 参数为尽力提供，可能不存在
}

type ResourceProviderManager interface {
	ResourceProviderFinder
	RegisterResource(int, string, ResourceProvider)
}

type ResourceProviderManagerImpl struct {
	ResourceProviderManager
}

type ResourceProviderFinder interface {
	QueryResource(resource *Resource) (ResourceProvider, error)
}

type ResourceProviderFinderImpl struct {
	Finder ResourceProviderFinder
}

func (b *ResourceProviderFinderImpl) ResourceMeta(resource *Resource) (*Metadata, error) {
	provider, err := b.Finder.QueryResource(resource)
	if err != nil {
		return nil, err
	}
	return provider.Metadata(resource.Scheme, resource.Body)
}

func (b *ResourceProviderFinderImpl) ResourceBlob(resource *Resource) (io.ReadCloser, error) {
	provider, err := b.Finder.QueryResource(resource)
	if err != nil {
		return nil, err
	}
	return provider.Reader(resource.Scheme, resource.Body)
}

type ResourceProvider interface {
	Metadata(scheme, body string) (*Metadata, error)
	Reader(scheme, body string) (io.ReadCloser, error)
}
