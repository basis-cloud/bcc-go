package bcc

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func convertTagsToNames(tags []Tag) []string {
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}
	return tagNames
}
