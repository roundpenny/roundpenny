output "repository_urls" {
  value = {
    for k, repo in aws_ecr_repository.service : k => repo.repository_url
  }
}

output "repository_arns" {
  value = {
    for k, repo in aws_ecr_repository.service : k => repo.arn
  }
}
