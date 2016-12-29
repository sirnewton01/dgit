package git

import (
	"strings"
)

// A RefSpec refers to a reference contained under .git/refs
type RefSpec string

func (r RefSpec) String() string {
	if len(r) < 1 {
		return ""
	}

	// This will only trim a single nil byte, but if there's more
	// than that we're doing something really wrong.
	return strings.TrimSpace(strings.TrimSuffix(string(r), "\000"))
}

// Returns the file that holds r.
func (r RefSpec) File(c *Client) File {
	return c.GitDir.File(File(r.String()))
}

// Returns the value of RefSpec in Client's GitDir, or the empty string
// if it doesn't exist.
func (r RefSpec) Value(c *Client) (string, error) {
	f := r.File(c)
	val, err := f.ReadAll()
	return strings.TrimSpace(val), err
}

// A Branch is a type of RefSpec that lives under refs/heads/ or refs/remotes/heads
// Use GetBranch to get a valid branch from a branchname, don't cast from string
type Branch RefSpec

// Returns a valid Branch object for an existing branch.
func GetBranch(c *Client, branchname string) (Branch, error) {
	b := Branch("refs/heads/" + branchname)
	if !b.Exists(c) {
		return "", InvalidBranch
	}
	return b, nil
}

// Returns true if the branch exists under c's GitDir
func (b Branch) Exists(c *Client) bool {
	return c.GitDir.File(File(b)).Exists()
}

// Implements Commitish interface on Branch.
func (b Branch) CommitID(c *Client) (CommitID, error) {
	val, err := RefSpec(b).Value(c)
	if err != nil {
		return CommitID{}, err
	}
	sha, err := Sha1FromString(val)
	return CommitID(sha), err
}

// Implements Treeish on Branch.
func (b Branch) TreeID(c *Client) (TreeID, error) {
	cmt, err := b.CommitID(c)
	if err != nil {
		return TreeID{}, err
	}
	return cmt.TreeID(c)
}

// Returns the branch name, without the refspec portion.
func (b Branch) BranchName() string {
	return strings.TrimPrefix(string(b), "refs/heads/")
}