package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"code.gitea.io/sdk/gitea"
	log "github.com/sirupsen/logrus"

	"github.com/xanzy/go-gitlab"
)

type gitlabMigrator struct {
	gl      *gitlab.Client
	gt      *gitea.Client
	gtOrgID int
}

var (
	debug          bool
	gitlabUsername string
	gitlabToken    string
	gitlabEndpoint string
	gitlabGroup    string
	giteaEndpoint  string
	giteaToken     string
	giteaOrg       string
	glm            gitlabMigrator

	commit  string
	builtAt string
	builtBy string
	builtOn string
)

func init() {
	flag.StringVar(&gitlabUsername, "gitlab-username", os.Getenv("GITLAB_USERNAME"), "Gitlab username")
	flag.StringVar(&gitlabToken, "gitlab-token", os.Getenv("GITLAB_TOKEN"), "Gitlab token")
	flag.StringVar(&gitlabEndpoint, "gitlab-endpoint", os.Getenv("GITLAB_ENDPOINT"), "Gitlab endpoint")
	flag.StringVar(&gitlabGroup, "gitlab-group", os.Getenv("GITLAB_GROUP"), "Gitlab group to clone repos frome")
	flag.StringVar(&giteaToken, "gitea-token", os.Getenv("GITEA_TOKEN"), "Gitea token")
	flag.StringVar(&giteaEndpoint, "gitea-endpoint", os.Getenv("GITEA_ENDPOINT"), "Gitea endpoint")
	flag.StringVar(&giteaOrg, "gitea-org", os.Getenv("GITEA_ORG"), "Gitea organisation")
	flag.BoolVar(&debug, "debug", isEnvDefined("DEBUG"), "enable debug output")

	flag.Parse()
}

func isEnvDefined(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func main() {
	fmt.Print("Version info :: ")
	fmt.Printf("commit: %s ", commit)
	fmt.Printf("built @ %s by %s on %s\n", builtAt, builtBy, builtOn)

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// Sanity checks
	if gitlabUsername == "" || gitlabToken == "" || gitlabEndpoint == "" || gitlabGroup == "" || giteaEndpoint == "" || giteaToken == "" || giteaOrg == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// gitlab client
	lab, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabEndpoint))
	if err != nil {
		log.WithField("err", err).Fatal("Failed to create Gitlab client")
	}

	// gitea client
	tea := gitea.NewClient(giteaEndpoint, giteaToken)
	glm = gitlabMigrator{
		gl: lab,
		gt: tea,
	}

	// Populate destination Gitea Organisation ID
	if glm.gtOrgID, err = glm.getGiteaOrganisationID(giteaOrg); err != nil {
		log.WithError(err).Fatal("Unable to fetch Gitea organisation")
	}

	projects, err := glm.fetchGitlabProjects(gitlabGroup)

	for _, p := range projects {
		glm.createGiteaMigration(p)
	}
}

func (glm *gitlabMigrator) fetchGitlabProjects(group string) ([]*gitlab.Project, error) {
	listProjects := []*gitlab.Project{}

	groups, res, err := glm.gl.Groups.SearchGroup(group)

	if err != nil {
		log.WithFields(
			log.Fields{
				"err":   err,
				"group": group,
			}).Error("Unable to search for group")
		return nil, err
	}
	if res.TotalItems != 1 {
		log.WithField("count", res.TotalItems).Fatal("Could not get exact group match")
		return nil, fmt.Errorf("Could not get exact group match (found %d", res.TotalItems)
	}

	for _, g := range groups {
		projects, _, err := glm.gl.Groups.ListGroupProjects(g.ID, nil)
		if err != nil {
			log.WithFields(
				log.Fields{
					"err":   err,
					"group": group,
				}).Error("Unable to list projects")
		}
		for _, p := range projects {
			log.WithField("url", p.HTTPURLToRepo).Debug("Adding repository...")
			listProjects = append(listProjects, p)
		}
	}

	return listProjects, nil
}

func (glm *gitlabMigrator) getGiteaOrganisationID(org string) (int, error) {
	teaOrg, err := glm.gt.GetOrg(org)
	if err != nil {
		return 0, err
	}
	return int(teaOrg.ID), nil
}

func (glm *gitlabMigrator) createGiteaMigration(remote *gitlab.Project) error {

	_, err := glm.gt.GetMyUserInfo() // Test Credentials
	if err != nil {
		return err
	}

	_, err = glm.gt.MigrateRepo(gitea.MigrateRepoOption{
		CloneAddr:    remote.HTTPURLToRepo,
		AuthUsername: gitlabUsername,
		AuthPassword: gitlabToken,
		UID:          glm.gtOrgID,
		RepoName:     remote.Name,
		Description:  remote.Description,
		Mirror:       true,
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "409") {
			log.WithField("src", remote.HTTPURLToRepo).Debug("Repository is already mirrored")
		} else {
			log.WithError(err).Error("Unable to clone repository")
		}
		return err
	}

	return nil
}
