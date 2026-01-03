package github

import "time"

// Repository is a GitHub repository response
type Repository struct {
	FullName         string    `json:"full_name"`
	Description      string    `json:"description"`
	StargazersCount  int       `json:"stargazers_count"`
	ForksCount       int       `json:"forks_count"`
	OpenIssuesCount  int       `json:"open_issues_count"`
	SubscribersCount int       `json:"subscribers_count"`
	PushedAt         time.Time `json:"pushed_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	License          *License  `json:"license"`
	Archived         bool      `json:"archived"`
	Disabled         bool      `json:"disabled"`
}

// License is a GitHub license
type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
}

// Commit is a GitHub commit
type Commit struct {
	SHA    string       `json:"sha"`
	Commit CommitDetail `json:"commit"`
}

// CommitDetail contains commit metadata
type CommitDetail struct {
	Author    CommitAuthor `json:"author"`
	Committer CommitAuthor `json:"committer"`
	Message   string       `json:"message"`
}

// CommitAuthor is a commit author
type CommitAuthor struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}
