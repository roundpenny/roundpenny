resource "aws_db_subnet_group" "this" {
  name       = "roundup-platform-${var.environment}"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "roundup-platform-${var.environment}"
  }
}

resource "aws_security_group" "rds" {
  name        = "roundup-platform-${var.environment}-rds"
  description = "RDS security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
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
    Name = "roundup-platform-${var.environment}-rds"
  }
}

resource "aws_db_parameter_group" "this" {
  name   = "roundup-platform-${var.environment}"
  family = "postgres16"

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "random_page_cost"
    value = "1.1"
  }

  parameter {
    name  = "effective_cache_size"
    value = "{DBInstanceClassMemory*32768}"
  }
}

resource "aws_db_instance" "primary" {
  identifier = "roundup-platform-${var.environment}"

  engine         = "postgres"
  engine_version = "16.3"
  instance_class = var.instance_class

  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = "roundup_platform"
  username = var.username
  password = var.password

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  backup_retention_period = var.backup_retention_days
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"

  multi_az = var.multi_az

  copy_tags_to_snapshot     = true
  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "roundup-platform-${var.environment}-final-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"

  performance_insights_enabled          = true
  performance_insights_retention_period = 7

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  tags = {
    Name = "roundup-platform-${var.environment}"
  }
}

resource "aws_db_instance" "read_replica" {
  count = var.multi_az ? 0 : 1

  identifier = "roundup-platform-${var.environment}-reader"

  instance_class = var.instance_class

  replicate_source_db = aws_db_instance.primary.identifier

  vpc_security_group_ids = [aws_security_group.rds.id]

  backup_retention_period = var.backup_retention_days
  backup_window           = "04:00-05:00"
  maintenance_window      = "sun:05:00-sun:06:00"

  copy_tags_to_snapshot = true
  deletion_protection   = true
  skip_final_snapshot   = false

  performance_insights_enabled          = true
  performance_insights_retention_period = 7

  enabled_cloudwatch_logs_exports = ["postgresql"]

  tags = {
    Name = "roundup-platform-${var.environment}-reader"
  }
}
