resource "aws_security_group" "msk" {
  name        = "roundup-platform-${var.environment}-msk"
  description = "MSK security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 9092
    to_port         = 9098
    protocol        = "tcp"
    security_groups = [var.eks_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "roundup-platform-${var.environment}-msk"
  }
}

resource "aws_msk_cluster" "this" {
  cluster_name           = "roundup-platform-${var.environment}"
  kafka_version          = "3.6.0"
  number_of_broker_nodes = var.broker_count

  broker_node_group_info {
    instance_type   = var.broker_instance_type
    client_subnets  = var.private_subnet_ids
    security_groups = [aws_security_group.msk.id]

    storage_info {
      ebs_storage_info {
        volume_size = var.ebs_volume_size
      }
    }
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.msk.arn

    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  client_authentication {
    unauthenticated = true
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.msk.name
      }
    }
  }

  tags = {
    Name = "roundup-platform-${var.environment}"
  }
}

resource "aws_kms_key" "msk" {
  description             = "MSK encryption key"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_cloudwatch_log_group" "msk" {
  name              = "/aws/msk/roundup-platform-${var.environment}"
  retention_in_days = 30
}

resource "aws_sns_topic" "msk_alerts" {
  name = "roundup-platform-${var.environment}-msk-alerts"
}

resource "aws_sns_topic_subscription" "msk_alerts_email" {
  count     = var.alert_email != "" ? 1 : 0
  topic_arn = aws_sns_topic.msk_alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}
