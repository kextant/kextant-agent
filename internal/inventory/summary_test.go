package inventory

import (
	"testing"

	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSummarizeComponentsDeduplicatesImagesAndPreservesLocations(t *testing.T) {
	summaries := summarizeComponents([]types.ComponentInstance{
		{Image: "repo/app:v1", ImageRepository: "repo/app", ImageTag: "v1", Namespace: "b", Kind: "Deployment", Name: "api", ContainerName: "app", ContainerType: "container"},
		{Image: "repo/app:v1", ImageRepository: "repo/app", ImageTag: "v1", Namespace: "a", Kind: "Job", Name: "worker", ContainerName: "app", ContainerType: "container"},
		{Image: "repo/other:v2", ImageRepository: "repo/other", ImageTag: "v2", Namespace: "a", Kind: "Pod", Name: "other", ContainerName: "other", ContainerType: "container"},
	})

	require.Len(t, summaries, 2)
	require.Equal(t, "repo/app:v1", summaries[0].Image)
	require.Len(t, summaries[0].Locations, 2)
	require.Equal(t, "a", summaries[0].Locations[0].Namespace)
	require.Equal(t, "b", summaries[0].Locations[1].Namespace)
	require.Equal(t, "repo/other:v2", summaries[1].Image)
}
