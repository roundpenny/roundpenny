resource "aws_security_group" "eks_cluster" {
  name        = "roundup-platform-${var.environment}-eks-cluster-sg"
  description = "Additional security group for EKS cluster"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [var.alb_security_group_id]
    description     = "Kong HTTPS"
  }

  ingress {
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [var.alb_security_group_id]
    description     = "Kong HTTP"
  }

  ingress {
    from_port   = 1024
    to_port     = 65535
    protocol    = "tcp"
    self        = true
    description = "Node-to-node communication"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "roundup-platform-${var.environment}-eks-cluster-sg"
  }
}

resource "aws_security_group_rule" "cluster_ingress_self" {
  security_group_id = aws_security_group.eks_cluster.id
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  self              = true
  description       = "Kubernetes API server"
}
