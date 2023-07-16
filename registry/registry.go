package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/stenstromen/ncregistry/types"
)

func GetRepositories(url, username, password string) (types.RepositoryResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/_catalog", nil)
	if err != nil {
		return types.RepositoryResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.RepositoryResponse{}, err
	}
	defer resp.Body.Close()

	var repoResp types.RepositoryResponse
	err = json.NewDecoder(resp.Body).Decode(&repoResp)
	return repoResp, err
}

func GetTags(url, username, password, repository string) (types.TagResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/tags/list", nil)
	if err != nil {
		return types.TagResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.TagResponse{}, err
	}
	defer resp.Body.Close()

	var tagResp types.TagResponse
	err = json.NewDecoder(resp.Body).Decode(&tagResp)
	return tagResp, err
}

func GetManifest(url, username, password, repository, tag string) (*types.ManifestResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/manifests/"+tag, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifestResp types.ManifestResponse
	err = json.NewDecoder(resp.Body).Decode(&manifestResp)
	if err != nil {
		return nil, err
	}

	return &manifestResp, nil
}

func GetBlob(url, username, password, repository, digest string) (types.BlobResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/blobs/"+digest, nil)
	if err != nil {
		return types.BlobResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.BlobResponse{}, err
	}
	defer resp.Body.Close()

	var blobResp types.BlobResponse
	err = json.NewDecoder(resp.Body).Decode(&blobResp)
	return blobResp, err
}

func DeleteManifest(url, username, password, repository, digest string) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url+"/v2/"+repository+"/manifests/"+digest, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return fmt.Errorf("unexpected response from server: %s, body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func DockerPull(url, repository, tag string) error {
	cmd := exec.Command("docker", "pull", url+"/"+repository+":"+tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
