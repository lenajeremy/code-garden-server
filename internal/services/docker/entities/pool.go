package entities

import "github.com/docker/docker/api/types"

type ContainerPool struct {
	Language
	Capacity   int
	Containers []types.Container
}

func NewContainerPool(language Language, capacity int) *ContainerPool {
	p := &ContainerPool{
		language,
		capacity,
		[]types.Container{},
	}
	return p
}

func (p *ContainerPool) StartContainers() error {
	image := LanguageImageMap[p.Language]
	for i := 0; i < p.Capacity; i++ {
		container := StartContainer(image)
		p.Containers = append(p.Containers, container)
	}
	return nil
}

func StartContainer(image string) types.Container {

}
