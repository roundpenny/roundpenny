# SSL/TLS Automation

Three options are available for automating SSL/TLS certificate management in the RoundPenny platform.

## Option 1: Let's Encrypt + Certbot (Docker Compose)

For development or simpler deployments using Docker Compose.

**Setup:**
```bash
# Obtain initial certificates
./deploy/ssl/init-letsencrypt.sh app.roundpenny.com admin@roundpenny.com

# Certbot auto-renews via the certbot-compose.yml service
docker compose -f deploy/ssl/certbot-compose.yml up -d
```

The `certbot-compose.yml` service runs a renewal loop every 12 hours. Certificates are stored in `deploy/ssl/data/certbot/conf/`.

**Webroot challenge** requires a web server to serve `/var/www/certbot` on port 80. Add this volume mount to your reverse proxy (e.g., Kong or Nginx) to complete the ACME challenge.

---

## Option 2: Kong SSL Termination

Kong already has TLS configuration in `deploy/kong/kong.yml`. To terminate SSL at the Kong gateway:

1. Place your TLS certificate and key files on the Kong container or mount them as secrets.
2. Configure the Kong service to listen on port 443 with SSL:
   ```yaml
   # Add to deploy/kong/kong.yml or use Kong Admin API
   services:
     - name: auth-service
       url: http://auth-service:8081
       protocol: https
       # ...
   ```
3. With cert-manager in Kubernetes, Kong ingress automatically provisions Let's Encrypt certificates.

---

## Option 3: Helm cert-manager (Kubernetes)

For production Kubernetes deployments, cert-manager automates certificate lifecycle.

**Prerequisites:**
```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade -i cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true
```

**Deploy ClusterIssuer:**
```bash
kubectl apply -f deploy/ssl/letsencrypt-clusterissuer.yaml
```

**Annotate Ingress resources** to automatically provision certificates:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: roundup-platform-ingress
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    kubernetes.io/tls-acme: "true"
spec:
  tls:
    - hosts:
        - app.roundpenny.com
      secretName: roundpenny-tls
  rules:
    - host: app.roundpenny.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kong-gateway
                port:
                  number: 80
```

**Certificate renewal** is handled automatically by cert-manager (typically 30 days before expiry).
