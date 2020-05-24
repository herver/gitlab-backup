# gitlab-backup

## Use case

This small utility is meant to be run via `cron` or `systemd` timers to
backup periodically `git` repositories hosted on a gitlab instance and
belonging to a specific group

## Usage

You can pass arguments via the CLI or by setting environment variables:

| Usage | flag | env var|
|-------|------|--------|
| gitlab-username | GITLAB_USERNAME | Gitlab username |
| gitlab-token | GITLAB_TOKEN | Gitlab token |
| gitlab-endpoint | GITLAB_ENDPOINT | Gitlab URL |
| gitlab-group | GITLAB_GROUP | Gitlab group to clone repos from |
| backup-dir | BACKUP_DIR | Where to write clones |

## API Token creation

The tool requires the API token with the following scopes:

* `api`
* `read_repository`

It can be created by going to `<YOUR_GITLAB_INSTALLATION>/profile/personal_access_tokens`
