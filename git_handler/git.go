package git_handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.uber.org/zap"

	"rmk/system"
)

const (
	PrefixFeature     = "feature/"
	PrefixRelease     = "release/"
	DefaultDevelop    = "develop"
	DefaultStaging    = "staging"
	DefaultProduction = "production"

	TaskNum int = iota
	SemVer
)

type GitSpec struct {
	DefaultBranches    []string
	DefaultBranch      string
	IntermediateBranch string
	RepoName           string
	RepoPrefixName     string
	ID                 string
	repo               *git.Repository
	auth               transport.AuthMethod
	workTree           *git.Worktree
	headRef            *plumbing.Reference
}

func (g *GitSpec) checkIntermediateBranchName(branch, prefix string) (int, error) {
	g.IntermediateBranch = strings.ReplaceAll(branch, prefix, "")
	patternTaskNum := regexp.MustCompile(`^[a-z]+-\d+`)
	patternSemVer := regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-z]+)?$`)

	switch {
	case len(patternTaskNum.FindString(strings.ToLower(g.IntermediateBranch))) > 0:
		g.IntermediateBranch = patternTaskNum.FindString(strings.ToLower(g.IntermediateBranch))
		return TaskNum, nil
	case len(patternSemVer.FindString(strings.ToLower(g.IntermediateBranch))) > 0:
		g.IntermediateBranch = strings.ReplaceAll(patternSemVer.FindString(strings.ToLower(g.IntermediateBranch)),
			".", "-")
		return SemVer, nil
	default:
		return 0, fmt.Errorf("selected branch %s cannot be used as environment name", branch)
	}
}

func (g *GitSpec) checkBranchName(branch string) error {
	for _, val := range g.DefaultBranches {
		if branch != val {
			switch {
			case strings.HasPrefix(branch, PrefixFeature):
				if _, err := g.checkIntermediateBranchName(branch, PrefixFeature); err != nil {
					return err
				}

				g.DefaultBranch = DefaultDevelop
			case strings.HasPrefix(branch, PrefixRelease):
				pattern, err := g.checkIntermediateBranchName(branch, PrefixRelease)
				if err != nil {
					return err
				}

				switch pattern {
				case TaskNum:
					g.DefaultBranch = DefaultStaging
				case SemVer:
					if strings.Contains(g.IntermediateBranch, "rc") {
						g.DefaultBranch = DefaultStaging
						break
					}

					g.DefaultBranch = DefaultProduction
				default:
					return fmt.Errorf("selected branch %s cannot be used as environment name", branch)
				}
			}
		} else {
			g.DefaultBranch = val
		}
	}

	if len(g.DefaultBranch) == 0 && len(g.IntermediateBranch) == 0 {
		return fmt.Errorf("selected branch %s cannot be used as environment name", branch)
	}

	return nil
}

func (g *GitSpec) GetBranchName() error {
	openOptions := git.PlainOpenOptions{
		DetectDotGit: true,
	}

	repo, err := git.PlainOpenWithOptions(system.GetPwdPath(""), &openOptions)
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	if !head.Name().IsBranch() {
		return fmt.Errorf("it's not branch %s", head.Name().Short())
	}

	return g.checkBranchName(head.Name().Short())
}

func (g *GitSpec) GetRepoPrefix() error {
	openOptions := git.PlainOpenOptions{
		DetectDotGit: true,
	}

	repo, err := git.PlainOpenWithOptions(system.GetPwdPath(""), &openOptions)
	if err != nil {
		return err
	}

	c, err := repo.Config()
	if err != nil {
		return err
	}

	if _, ok := c.Remotes["origin"]; !ok {
		return fmt.Errorf("failed to extract prefix from repository name")
	} else {
		g.RepoName = strings.TrimSuffix(filepath.Base(strings.Join(c.Remotes["origin"].URLs, "")), ".git")

		if len(strings.Split(filepath.Base(g.RepoName), ".")) > 0 {
			g.RepoPrefixName = strings.Split(filepath.Base(g.RepoName), ".")[0]
		}

		return nil
	}
}

func (g *GitSpec) GenerateID() error {
	if err := g.GetBranchName(); err != nil {
		return err
	}

	if err := g.GetRepoPrefix(); err != nil {
		return err
	}

	if len(g.IntermediateBranch) > 0 {
		g.ID = g.RepoPrefixName + "-" + g.IntermediateBranch
	} else {
		g.ID = g.RepoPrefixName + "-" + g.DefaultBranch
	}

	return nil
}

func (g *GitSpec) GitCommitPush(pathRF, msg, token string) error {
	var err error

	if g.repo, err = git.PlainOpen(system.GetPwdPath("")); err != nil {
		return err
	}

	if pathRF, err = filepath.Rel(system.GetPwdPath(""), pathRF); err != nil {
		return err
	}

	if g.workTree, err = g.repo.Worktree(); err != nil {
		return err
	}

	if _, err := g.workTree.Add(pathRF); err != nil {
		return err
	}

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit Since version 5.0.1, we can omit the Author signature, being read
	// from the git config files.
	hash, err := g.workTree.Commit(msg, &git.CommitOptions{})
	if err != nil {
		return err
	}

	zap.S().Infof("hash commit - %s created with message: %s...", hash.String()[:7], strings.Split(msg, ",")[0])

	if g.headRef, err = g.repo.Head(); err != nil {
		return err
	}

	if g.auth, err = g.getAuthMethod(token); err != nil {
		return err
	}

	reset := &system.SpecCMD{
		Args:          []string{"reset", "--hard", "origin/" + g.headRef.Name().Short()},
		Command:       "git",
		Dir:           system.GetPwdPath(""),
		Ctx:           context.TODO(),
		DisableStdOut: true,
		Debug:         false,
	}

	cherryPick := &system.SpecCMD{
		Args:          []string{"cherry-pick", hash.String()},
		Command:       "git",
		Dir:           system.GetPwdPath(""),
		Ctx:           context.TODO(),
		DisableStdOut: true,
		Debug:         false,
	}

	fetchOpt := &git.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(
				fmt.Sprintf("+%s:refs/remotes/origin/%s",
					g.headRef.Name(),
					g.headRef.Name().Short(),
				),
			),
		},
		Auth:  g.auth,
		Tags:  git.NoTags,
		Force: true,
	}

	pushOpt := &git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(
				fmt.Sprintf("%s:refs/heads/%s",
					g.headRef.Name(),
					g.headRef.Name().Short(),
				),
			),
		},
		Auth:     g.auth,
		Progress: os.Stdout,
		Force:    false,
	}

	if err = g.repo.Push(pushOpt); err != nil {
		zap.S().Warnf("push operation: %s", err)

		zap.S().Infof("fetch remote origin: %s", g.headRef.Name().Short())
		if err := g.repo.Fetch(fetchOpt); err != nil {
			return err
		}

		zap.S().Infof("reset local branch %s by remote origin", g.headRef.Name().Short())
		if err := reset.ExecCMD(); err != nil {
			return fmt.Errorf("Git failed to reset local branch %s\n%s", g.headRef.Name().Short(),
				reset.StderrBuf.String())
		}

		zap.S().Infof("cherry-pick last hash commit %s", hash.String()[:7])
		if err := cherryPick.ExecCMD(); err != nil {
			return fmt.Errorf("Git failed to cherry-pick last hash commit %s\n%s", hash.String()[:7],
				cherryPick.StderrBuf.String())
		}

		if err := g.repo.Push(pushOpt); err != nil {
			return err
		}
	}

	return nil
}

func (g *GitSpec) getAuthMethod(token string) (transport.AuthMethod, error) {
	c, err := g.repo.Config()
	if err != nil {
		return nil, err
	}

	if _, ok := c.Remotes["origin"]; !ok {
		return nil, fmt.Errorf("failed to detect auth method")
	} else {
		urls := c.Remotes["origin"].URLs
		if len(urls) > 0 {
			if strings.Contains(urls[0], "http") {
				return &http.BasicAuth{Username: "git", Password: token}, nil
			}

			return ssh.NewPublicKeysFromFile("git", system.GetHomePath(system.GitSSHPrivateKey), "")
		}

		return nil, fmt.Errorf("failed to detect auth method")
	}
}
