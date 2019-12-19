package pkg

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)


func Checkout(gitUrl string, sha1 string, checkoutDirectory string, username string, password string) error {
	fmt.Println("Checkout: %v ==> %v",gitUrl, checkoutDirectory)
	
	repo, err := git.PlainClone(checkoutDirectory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL: gitUrl,
	})
	if err != nil {
		return err
	}
	
	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	hash := plumbing.NewHash(sha1)
	err = workTree.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
	if err != nil {
		return err
	}
	
	// ... retrieving the commit object to log informations
	if commit, err := repo.CommitObject(hash); err != nil {
		return err
	} else {
		fmt.Print(commit)
	}

	return nil
}






