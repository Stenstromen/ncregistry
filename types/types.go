package types

type Entry struct {
	URL      string
	Username string
	Password string
}

type Config struct {
	Entries []Entry
}

type RepositoryResponse struct {
	Repositories []string `json:"repositories"`
}

type TagResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ManifestResponse struct {
	Config struct {
		Digest string `json:"digest"`
		Size   int    `json:"size"`
	} `json:"config"`
	Layers []struct {
		Size int `json:"size"`
	} `json:"layers"`
}

type BlobResponse struct {
	Created string `json:"created"`
}

type TagInfo struct {
	Name string
	Date string
	Size int
}
