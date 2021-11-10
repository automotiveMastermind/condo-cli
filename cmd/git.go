package cmd

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

const (
	tag       string = `^refs/tags/(.+)$`
	header    string = `^(\w*)(?:\(([\w\$\.\-\* ]*)\))?\: (.*)`
	revert    string = `^Revert\s""([\s\S]*)""\s*This reverts commit (\w*)\.`
	breaking  string = `BREAKING CHANGES?`
	action    string = `Close(s|d)?|Fix(ed)?|Resolve(s|d)?`
	reference string = `#`
	mention   string = `@`
)

// Commit message struct
type Commit struct {
	Tag     string
	Version *semver.Version
	Hash    plumbing.Hash
	Type    string
	Scope   string
	Desc    string
	Author  object.Signature
	Actions map[string]string
}

func (c Commit) String() string {
	return fmt.Sprintf("[Hash: %s, Tag: %s,\tVersion: %s,\tType: %s,\tScope: %s,\tDesc: %s,\tAuthor: %s]", c.Hash.String()[0:7], c.Tag, c.Version, c.Type, c.Scope, c.Desc, c.Author)
}

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "run git",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gitRun()
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func gitRun() {
	r, _ := git.PlainOpen(".")

	tagRagex := regexp.MustCompile(tag)
	tagrefs, _ := r.Tags()
	tags := make(map[string]string)

	_ = tagrefs.ForEach(func(t *plumbing.Reference) error {
		m := tagRagex.FindAllStringSubmatch(t.Name().String(), 1)

		if m != nil {
			tags[t.Hash().String()] = m[0][1]
		}
		return nil
	})

	//... retrieving the HEAD reference
	ref, _ := r.Head()

	//... retrieves the commit history
	cIter, _ := r.Log(&git.LogOptions{From: ref.Hash()})
	headerRegex := regexp.MustCompile(header)
	var version semver.Version

	_ = cIter.ForEach(func(c *object.Commit) error {
		m := headerRegex.FindAllStringSubmatch(c.Message, -1)

		//fmt.Println(c.Message)
		if m != nil {
			tag := tags[c.Hash.String()]
			v, err := semver.StrictNewVersion(tag)
			if err == nil && v.GreaterThan(&version) {
				version = *v
			}

			c := Commit{
				Type:    m[0][1],
				Scope:   m[0][2],
				Desc:    m[0][3],
				Hash:    c.Hash,
				Tag:     tag,
				Version: v,
				Author:  c.Author,
			}
			fmt.Println(c)
		}

		return nil
	})

	fmt.Printf("Current Version: %s", version)
}
