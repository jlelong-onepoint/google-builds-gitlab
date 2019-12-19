package pkg

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)


func Checkout(gitUrl string, checkoutDirectory string, username string, password string) error {
	fmt.Printf("Checkout: %v ==> %v",gitUrl, checkoutDirectory)
	
	git, err := git.PlainClone(checkoutDirectory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL: gitUrl,
	})
	if err != nil {
		return err
	}

	// ... retrieving the branch being pointed by HEAD
	ref, err := git.Head()
	if err != nil {
		return err
	}
	
	// ... retrieving the commit object
	if commit, err := git.CommitObject(ref.Hash()); err != nil {
		return err
	} else {
		fmt.Print(commit)
	}

	return nil
}






