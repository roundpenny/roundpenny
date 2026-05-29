locals {
  services = [
    "auth-service",
    "transaction-service",
    "wallet-service",
    "fee-service",
    "investment-service",
    "payment-gateway",
    "ledger-service",
    "roundup-engine",
    "user-service",
    "notification-service",
    "merchant-service",
    "analytics-service",
    "fraud-service",
  ]
}

resource "aws_ecr_repository" "service" {
  for_each = toset(local.services)

  name                 = "roundup-platform/${each.key}"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
  }

  tags = {
    Name = "roundup-platform-${each.key}"
  }
}

resource "aws_ecr_lifecycle_policy" "service" {
  for_each = aws_ecr_repository.service

  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Expire untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
