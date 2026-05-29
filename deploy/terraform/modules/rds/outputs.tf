output "endpoint" {
  value = aws_db_instance.primary.endpoint
}

output "reader_endpoint" {
  value = var.multi_az ? aws_db_instance.primary.endpoint : try(aws_db_instance.read_replica[0].endpoint, "")
}

output "security_group_id" {
  value = aws_security_group.rds.id
}

output "db_name" {
  value = aws_db_instance.primary.db_name
}

output "db_username" {
  value = aws_db_instance.primary.username
}
