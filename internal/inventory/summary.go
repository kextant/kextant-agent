package inventory

import (
	"sort"

	"github.com/kextant/kextant-agent/pkg/types"
)

func summarizeComponents(components []types.ComponentInstance) []types.ComponentSummary {
	if len(components) == 0 {
		return nil
	}

	byImage := map[string]*types.ComponentSummary{}
	for _, component := range components {
		summary, ok := byImage[component.Image]
		if !ok {
			summary = &types.ComponentSummary{
				Image:           component.Image,
				ImageRegistry:   component.ImageRegistry,
				ImageRepository: component.ImageRepository,
				ImageTag:        component.ImageTag,
				ImageDigest:     component.ImageDigest,
			}
			byImage[component.Image] = summary
		}
		summary.Locations = append(summary.Locations, types.ComponentLocation{
			Namespace:     component.Namespace,
			Kind:          component.Kind,
			Name:          component.Name,
			ContainerName: component.ContainerName,
			ContainerType: component.ContainerType,
		})
	}

	images := make([]string, 0, len(byImage))
	for image := range byImage {
		images = append(images, image)
	}
	sort.Strings(images)

	summaries := make([]types.ComponentSummary, 0, len(images))
	for _, image := range images {
		summary := byImage[image]
		sort.Slice(summary.Locations, func(i, j int) bool {
			left := summary.Locations[i]
			right := summary.Locations[j]
			if left.Namespace != right.Namespace {
				return left.Namespace < right.Namespace
			}
			if left.Kind != right.Kind {
				return left.Kind < right.Kind
			}
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.ContainerName < right.ContainerName
		})
		summaries = append(summaries, *summary)
	}
	return summaries
}
