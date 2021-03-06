package version

import (
	"io/ioutil"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestGitDescribe(t *testing.T) {
	assert := assert.New(t)
	dir, _ := ioutil.TempDir("", "example")
	repo, err := git.PlainInit(dir, false)
	assert.NoError(err)

	worktree, err := repo.Worktree()
	assert.NoError(err)

	test := func(expected *RepoHead) {
		actual, err := GitDescribe(dir)
		assert.NoError(err)
		assert.Equal(expected, actual)
	}

	author := &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
	}
	opts := git.CommitOptions{Author: author}

	commit1, err := worktree.Commit("first commit", &opts)
	assert.NoError(err)
	test(&RepoHead{Hash: commit1.String(), CommitsSinceTag: 1})

	tag1, err := repo.CreateTag("1.0.0", commit1, nil)
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1.Name().Short(),
		Hash:            commit1.String(),
		CommitsSinceTag: 0,
	})

	tag1Post, err := repo.CreateTag("v1.0.0", commit1, &git.CreateTagOptions{
		Tagger:  author,
		Message: "annotated tag revisited",
	})
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1Post.Name().Short(),
		Hash:            commit1.String(),
		CommitsSinceTag: 0,
	})

	commit2, err := worktree.Commit("second commit", &opts)
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1Post.Name().Short(),
		Hash:            commit2.String(),
		CommitsSinceTag: 1,
	})

	tag2, err := repo.CreateTag("v2.0.0-rc.1", commit2, &git.CreateTagOptions{
		Tagger:  author,
		Message: "annotated tag",
	})
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag2.Name().Short(),
		Hash:            commit2.String(),
		CommitsSinceTag: 0,
	})
}

func TestGitDescribeError(t *testing.T) {
	assert := assert.New(t)
	dir, _ := ioutil.TempDir("", "example")

	test := func(msg string) {
		head, err := GitDescribe(dir)
		assert.Nil(head)
		assert.EqualError(err, msg)
	}
	test("failed to open repo: repository does not exist")

	_, err := git.PlainInit(dir, false)
	assert.NoError(err)
	test("failed to retrieve repo head: reference not found")
}
