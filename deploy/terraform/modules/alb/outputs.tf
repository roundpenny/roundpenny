output "alb_id" {
  value = aws_lb.kong.id
}

output "kong_alb_dns" {
  value = aws_lb.kong.dns_name
}

output "alb_security_group_id" {
  value = aws_security_group.alb.id
}
