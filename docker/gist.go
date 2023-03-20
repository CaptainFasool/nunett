package docker

// This files keeps all the functions related to Gist communication

import (
	"bytes"
	"errors"
	"log"

	"github.com/google/go-github/github"
)

func createGist() (*github.Gist, *github.Response, error) {
	gist := &github.Gist{
		Description: github.String("Docker Logs to Gist"),
		Public:      github.Bool(false),
		Files: map[github.GistFilename]github.GistFile{
			"stdout.log": {Content: github.String("No updates from docker container to stdout stream")},
			"stderr.log": {Content: github.String("No updates from docker container to stderr stream")},
		},
	}

	createdGist, resp, err := gh.Gists.Create(ctx, gist)
	if err != nil {
		return createdGist, resp, err
	}

	log.Println(*createdGist.HTMLURL)
	log.Println("[gist]: Remaining request quota:", resp.Remaining) // if this is equal to 0, we have exhausted limit for 24 hours.
	if resp.Remaining == 0 {
		return createdGist, resp, errors.New("gist quota exhausted")
	}

	return createdGist, resp, err
}

func updateGist(gistID string, stdout bytes.Buffer, stderr bytes.Buffer) {
	var errGistFile github.GistFile
	var outGistFile github.GistFile

	if stderr.String() == "" {
		errGistFile = github.GistFile{Content: github.String("No updates from docker container to stderr stream")}
	} else {
		errGistFile = github.GistFile{Content: github.String(cleanFlushInfo(&stderr))}
	}
	if stdout.String() == "" {
		outGistFile = github.GistFile{Content: github.String("No updates from docker container to stdout stream")}
	} else {
		outGistFile = github.GistFile{Content: github.String(cleanFlushInfo(&stdout))}
	}

	gist := &github.Gist{
		Files: map[github.GistFilename]github.GistFile{
			"stdout.log": outGistFile,
			"stderr.log": errGistFile,
		},
	}
	editedGist, resp, err := gh.Gists.Edit(ctx, gistID, gist)
	if err != nil {
		panic(err)
	}

	log.Println("UpdatedAt:", editedGist.GetUpdatedAt())
	log.Println("Resp Code:", resp.StatusCode) // if this is equal to 0, we have exhausted limit for 24 hours.
	// log.Printf("%v\n", resp.Header)
}
