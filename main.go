package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

var (
	debug          bool
	GitlabUsername string
	GitlabToken    string
	GitlabEndpoint string
	GitlabGroup    string
	BackupDir      string
)

func init() {
	flag.StringVar(&GitlabUsername, "gitlab-username", os.Getenv("GITLAB_USERNAME"), "Gitlab username")
	flag.StringVar(&GitlabToken, "gitlab-token", os.Getenv("GITLAB_TOKEN"), "Gitlab token")
	flag.StringVar(&GitlabEndpoint, "gitlab-endpoint", os.Getenv("GITLAB_ENDPOINT"), "Gitlab endpoint")
	flag.StringVar(&GitlabGroup, "gitlab-group", os.Getenv("GITLAB_GROUP"), "Gitlab group to clone repos frome")
	flag.StringVar(&BackupDir, "backup-dir", os.Getenv("BACKUP_DIR"), "Where to create the clones")
	flag.BoolVar(&debug, "debug", isEnvDefined("DEBUG"), "enable debug output")

	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// Sanity checks
	if GitlabUsername == "" || GitlabToken == "" || GitlabEndpoint == "" || GitlabGroup == "" || BackupDir == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func isEnvDefined(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func main() {
	git, err := gitlab.NewClient(GitlabToken, gitlab.WithBaseURL(GitlabEndpoint))
	if err != nil {
		log.WithField("err", err).Fatal("Failed to create client")
	}

	groups, res, err := git.Groups.SearchGroup(GitlabGroup)
	if err != nil {
		log.WithFields(
			log.Fields{
				"err":   err,
				"group": GitlabGroup,
			}).Fatal("Unable to search for group")
	}
	if res.TotalItems != 1 {
		log.WithField("count", res.TotalItems).Fatal("Could not get exact group match")
	}

	for _, g := range groups {
		projects, _, err := git.Groups.ListGroupProjects(g.ID, nil)
		if err != nil {
			log.WithFields(
				log.Fields{
					"err":   err,
					"group": GitlabGroup,
				}).Fatal("Unable to list projects")
		}
		for _, p := range projects {
			log.WithField("url", p.HTTPURLToRepo).Info("Cloning repository...")
			err := backupGitRepo(p.HTTPURLToRepo)
			if err != nil {
				log.WithField("err", err).Error("Unable to clone GIT repo")
				continue
			}
		}

	}
}

func backupGitRepo(remote string) error {
	// Sanity checks
	u, err := url.Parse(remote)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%s is not supported at the moment", u.Scheme)
	}

	destPath := path.Join(BackupDir, u.Path)

	// If destPath isn't a git repo, clone
	if _, err := os.Stat(path.Join(destPath, ".git")); os.IsNotExist(err) {
		log.WithField("repo", remote).Info("Cloning new repository")
		_, err = git.PlainClone(destPath, false, &git.CloneOptions{
			URL: remote,
			Auth: &http.BasicAuth{
				Username: GitlabUsername,
				Password: GitlabToken,
			},
		})

		if err != nil {
			return err
		}
	} else {
		// Otherwise just git pull
		repo, err := git.PlainOpen(destPath)
		log.Debug("Opening repository")
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		log.Debug("Opening worktree")
		if err != nil {
			return err
		}

		log.Debug("Pulling from default remote")
		w.Pull(&git.PullOptions{
			RemoteName: "origin",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
