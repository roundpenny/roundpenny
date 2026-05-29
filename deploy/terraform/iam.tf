data "aws_iam_policy_document" "eks_service_account" {
  statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = ["*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "kms:Decrypt",
      "kms:DescribeKey",
    ]

    resources = ["*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]

    resources = ["*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "kafka:DescribeCluster",
      "kafka:GetBootstrapBrokers",
      "kafka:DescribeClusterV2",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role" "service_account" {
  name = "roundup-platform-${var.environment}-service-account"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = module.eks.cluster_oidc_issuer_url
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(module.eks.cluster_oidc_issuer_url, "https://", "")}:sub" = "system:serviceaccount:default:roundup-platform"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "service_account" {
  name   = "roundup-platform-${var.environment}-service-account"
  role   = aws_iam_role.service_account.name
  policy = data.aws_iam_policy_document.eks_service_account.json
}
