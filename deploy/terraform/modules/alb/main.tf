resource "aws_security_group" "alb" {
  name        = "roundup-platform-${var.environment}-alb"
  description = "ALB security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from anywhere"
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from anywhere"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "roundup-platform-${var.environment}-alb"
  }
}

resource "aws_lb" "kong" {
  name               = "roundup-platform-${var.environment}"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = true
  idle_timeout               = 60

  tags = {
    Name = "roundup-platform-${var.environment}"
  }
}

resource "aws_lb_target_group" "kong_http" {
  name     = "roundup-platform-${var.environment}-http"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/v1/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 10
    timeout             = 5
    matcher             = "200"
  }

  tags = {
    Name = "roundup-platform-${var.environment}-http"
  }
}

resource "aws_lb_target_group" "kong_https" {
  name     = "roundup-platform-${var.environment}-https"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = var.vpc_id

  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/v1/health"
    port                = "traffic-port"
    protocol            = "HTTPS"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 10
    timeout             = 5
    matcher             = "200"
  }

  tags = {
    Name = "roundup-platform-${var.environment}-https"
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.kong.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.kong.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.kong_https.arn
  }
}

resource "aws_lb_target_group_attachment" "kong_http" {
  for_each = var.kong_private_ips

  target_group_arn = aws_lb_target_group.kong_http.arn
  target_id        = each.value
  port             = 80
}

resource "aws_lb_target_group_attachment" "kong_https" {
  for_each = var.kong_private_ips

  target_group_arn = aws_lb_target_group.kong_https.arn
  target_id        = each.value
  port             = 443
}
