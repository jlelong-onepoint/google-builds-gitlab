package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)


func Checkout(gitUrl string, sha1 string, checkoutDirectory string, username string, password string) {
	fmt.Printf("Checkout: %v ==> %v\n",gitUrl, checkoutDirectory)
	
	repo, err := git.PlainClone(checkoutDirectory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL: gitUrl,
	})
	if err != nil {
		panic(err)
	}
	
	workTree, err := repo.Worktree()
	if err != nil {
		panic(err)
	}

	hash := plumbing.NewHash(sha1)
	err = workTree.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
	if err != nil {
		panic(errors.Wrapf(err, "Unable to checkout sha %v", sha1))
	}
	
	// ... retrieving the commit object to log informations
	if commit, err := repo.CommitObject(hash); err != nil {
		panic(err)
	} else {
		fmt.Println(commit)
	}
}






