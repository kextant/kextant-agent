package inventory

import "strings"

type parsedImage struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string
}

func parseImageReference(image string) parsedImage {
	result := parsedImage{}
	withoutDigest := image
	if before, after, ok := strings.Cut(image, "@"); ok {
		withoutDigest = before
		result.Digest = after
	}

	lastSlash := strings.LastIndex(withoutDigest, "/")
	lastColon := strings.LastIndex(withoutDigest, ":")
	withoutTag := withoutDigest
	if lastColon > lastSlash {
		withoutTag = withoutDigest[:lastColon]
		result.Tag = withoutDigest[lastColon+1:]
	}

	parts := strings.Split(withoutTag, "/")
	if len(parts) > 1 && isRegistry(parts[0]) {
		result.Registry = parts[0]
		result.Repository = strings.Join(parts[1:], "/")
	} else {
		result.Repository = withoutTag
	}

	return result
}

func isRegistry(part string) bool {
	return strings.Contains(part, ".") || strings.Contains(part, ":") || part == "localhost"
}
