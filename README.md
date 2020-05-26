# gitlab-backup

![](https://github.com/herver/gitlab-backup/workflows/Build/badge.svg)

## Use case

This small utility is meant to be run via `cron` or `systemd` timers to
create migration (aka mirroring) tasks between a source (Gitlab) and a destination (Gitea)
belonging to a specific group

## Usage

You can pass arguments via the CLI or by setting environment variables:

| Usage | flag | env var|
|-------|------|--------|
| gitlab-username | GITLAB_USERNAME | Gitlab username |
| gitlab-token | GITLAB_TOKEN | Gitlab token |
| gitlab-endpoint | GITLAB_ENDPOINT | Gitlab URL |
| gitlab-group | GITLAB_GROUP | Gitlab group to clone repos from |
| gitea-token | GITEA_TOKEN | Gitea token |
| gitea-endpoint | GITEA_ENDPOINT | Gitea URL |
| gitea-org| GITEA_ORG | Gitea organisation to mirror to |

## API Token creation

### Gitlab

The tool requires the API token with the following scopes:

* `api`
* `read_repository`

It can be created by going to `<YOUR_GITLAB_INSTALLATION>/profile/personal_access_tokens`

### Gitea

A standard Gitea Application token is used by the application, it can be created by going to the following URL:

* `https://your-gitea.installation/user/settings/applications`

## SystemD integration

The tool comes with sample `service` and `timer` SystemD unit files to run gitlab-backup on a daily basis.

* Move the files in `/etc/systemd/system` (or wherever your distribution wants them)
* Enable the `timer` unit: `systemctl start gitlab-backup.timer`
* Check that is it enabled: `systemctl list-timers`
