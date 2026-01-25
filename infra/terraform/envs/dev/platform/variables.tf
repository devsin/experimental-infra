variable "region" {
  type    = string
  default = "europe-west2"
}

variable "git_repo_url" {
  type = string
}

variable "github_repository" {
  description = "GitHub repository in the format owner/repo"
  type        = string
}
