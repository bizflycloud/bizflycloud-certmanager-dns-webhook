# BizflyCloud Cert-manager DNS webhook

Cert-manager ACME DNS webhook provider for BizflyCloud DNS.

## Introduction

BizflyCloud Cert-manager DNS  is a webhook run in kubernetes to provide connect between cert-manager and Bizfly Cloud provider DNS.

## Why need to use BizflyCloud Cert-manager DNS webhook

As you know Let's encrypt use 2 method to provide certificate which is ACME HTTP01 and DNS01.

THis web hook will automaticly create DNS01 challenge solver in BizflyCloud DNS and apply the certificate to your ingress.

## Install

### Install cert manager

Install cert manager using this document here: <https://cert-manager.io/docs/installation/kubernetes/>

**Note**: If you customized the installation of cert-manager, you may need to also set the certManager.namespace and certManager.serviceAccountName values.

### Install webhook

#### Option 1

Install bizflycloud-certmanager-dns-webhook using helm

**Note**: Choose a unique group name to identify your company or organization (for example `acme.mycompany.example`).

Change your authentication value in `./deploy/bizflycloud-certmanager-dns-webhook/values.yaml`

```bash
helm install <deploy name> ./deploy/bizflycloud-certmanager-dns-webhook 
```

#### Option 2

Install bizflycloud-certmanager-dns-webhook using manifest.

**Notes**: Webhook's themselves are deployed as Kubernetes API services, in order to allow administrators to restrict access to webhooks with Kubernetes RBAC.

This is important, as otherwise it'd be possible for anyone with access to your webhook to complete ACME challenge validations and obtain certificates.

Install using the file `./manifest/bundle.yaml`

Change your groupname match ClusterIssuer in these deployment:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: bizflycloud-webhook:domain-solver
  labels:
    app: bizflycloud-webhook
rules:
  - apiGroups:
      - acme.mycompany.com
    resources:
      - '*'
    verbs:
      - 'create'
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bizflycloud-webhook
  namespace: cert-manager
  labels:
    app: bizflycloud-webhook
spec:
  selector:
    matchLabels:
      app: bizflycloud-webhook
  template:
    metadata:
      labels:
        app: bizflycloud-webhook
    spec:
      serviceAccountName: bizflycloud-webhook
      containers:
        - name: bizflycloud-webhook
          image: cr-hn-1.vccloud.vn/31ff9581861a4d0ea4df5e7dda0f665d/bizflycloud-certmanager-dns-webhook:latest
          imagePullPolicy: Always
          args:
            - --tls-cert-file=/tls/tls.crt
            - --tls-private-key-file=/tls/tls.key
          env:
            - name: GROUP_NAME
              value: "acme.mycompany.com"
```

```yaml
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1alpha1.acme.mycompany.com
  labels:
    app: bizflycloud-webhook
  annotations:
    cert-manager.io/inject-ca-from: "cert-manager/bizflycloud-webhook-webhook-tls"
spec:
  group: acme.mycompany.com
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: bizflycloud-webhook
    namespace: cert-manager
  version: v1alpha1
```

## Example

After install cert-manager and bizflycloud-certmanager-dns-webhook

1. Create 2 service for demo:

    echo1.yaml

    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
    name: echo1
    spec:
    ports:
    - port: 80
        targetPort: 5678
    selector:
        app: echo1
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: echo1
    spec:
    selector:
        matchLabels:
        app: echo1
    replicas: 2
    template:
        metadata:
        labels:
            app: echo1
        spec:
        containers:
        - name: echo1
            image: hashicorp/http-echo
            args:
            - "-text=echo1"
            ports:
            - containerPort: 5678
    ```

    echo2.yaml

    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
    name: echo2
    spec:
    ports:
    - port: 80
        targetPort: 5678
    selector:
        app: echo2
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: echo2
    spec:
    selector:
        matchLabels:
        app: echo2
    replicas: 1
    template:
        metadata:
        labels:
            app: echo2
        spec:
        containers:
        - name: echo2
            image: hashicorp/http-echo
            args:
            - "-text=echo2"
            ports:
            - containerPort: 5678
    ```

2. Install nginx-ingress-controller
    Follow this link: <https://engineering.bizflycloud.vn/cai-dat-nginx-ingress-controller-cho-kubernetes/>

    After that, use BizflyCloud DNS service to create record for your domain and sub-domain.

    the Ipv4 value is your Loadbalancer IP created by nginx-ingress above

    ![dns](https://raw.githubusercontent.com/lmq1999/123/master/image.png)

3. Create ClusterIssuer/Issuer

    ```yaml
    apiVersion: cert-manager.io/v1alpha2
    kind: ClusterIssuer
    metadata:
    name: letsencrypt-prod
    namespace: cert-manager
    spec:
    acme:
        # Change to your letsencrypt email
        email: example@example.com
        server: https://acme-v02.api.letsencrypt.org/directory
        privateKeySecretRef:
        name: letsencrypt-prod
        solvers:
        - dns01:
            webhook:
            groupName: acme.mycompany.com
            solverName: bizflycloud
    ```

4. Create nginx ingress

    ```yaml
    apiVersion: networking.k8s.io/v1beta1
    kind: Ingress
    metadata:
    name: echo-ingress
    annotations:
        cert-manager.io/cluster-issuer: "letsencrypt-prod"
    spec:
    tls:
    - hosts:
        - echo1.example.com
        - echo2.example.com
        secretName: echo-tls
    rules:
    - host: echo1.example.com
        http:
        paths:
        - backend:
            serviceName: echo1
            servicePort: 80
    - host: echo2.example.com
        http:
        paths:
        - backend:
            serviceName: echo2
            servicePort: 80
    ```

5. Wait a couple of minutes for the Let’s Encrypt production server to issue the certificate

6. Verify

```bash
quanlm@quanlm-desktop:~$ curl https://echo2.quanlm1999-testz.tk/
echo2
```

Using `curl -v` to see TLS handshake

```bash
quanlm@quanlm-desktop:~$ curl https://echo2.quanlm1999-testz.tk/ -v
*   Trying 14.225.0.197:443...
* TCP_NODELAY set
* Connected to echo2.quanlm1999-testz.tk (14.225.0.197) port 443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
* TLSv1.3 (IN), TLS handshake, Certificate (11):
* TLSv1.3 (IN), TLS handshake, CERT verify (15):
* TLSv1.3 (IN), TLS handshake, Finished (20):
* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.3 (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=echo1.quanlm1999-testz.tk
*  start date: Jul  8 03:30:37 2021 GMT
*  expire date: Oct  6 03:30:36 2021 GMT
*  subjectAltName: host "echo2.quanlm1999-testz.tk" matched cert's "echo2.quanlm1999-testz.tk"
*  issuer: C=US; O=Let's Encrypt; CN=R3
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x557cad710e10)
> GET / HTTP/2
> Host: echo2.quanlm1999-testz.tk
> user-agent: curl/7.68.0
> accept: */*
> 
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* old SSL session ID is stale, removing
* Connection state changed (MAX_CONCURRENT_STREAMS == 128)!
< HTTP/2 200 
< date: Thu, 08 Jul 2021 08:47:54 GMT
< content-type: text/plain; charset=utf-8
< content-length: 6
< x-app-name: http-echo
< x-app-version: 0.2.3
< strict-transport-security: max-age=15724800; includeSubDomains
< 
echo2
* Connection #0 to host echo2.quanlm1999-testz.tk left intact
```
