# BizflyCloud Cert-manager DNS webhook

## Giới thiệu

BizflyCloud Cert-manager DNS là một webhook thực hiện công việc xử lý DNS01 challenge với cert-manager

## Tại sao cần sử dụng BizflyCloud Cert-manager DNS webhook

Bizfly Cloud cung cấp dịch vụ DNS tại <https://manage.bizflycloud.vn/dns/>

Cert-manager khi tạo TLS/SSL certificate thì cần kiểm tra rằng bạn có đang sở hữu quyền quản lý tên miền đó không. Sử dụng phương thức Challenge DNS01, đó là gửi 1 key cho bạn và kiểm tra các bản ghi tên miền `_acme-challenge.<tên miền của bạn>` có giá trị là key vừa gửi.

Thay vì phải thủ công lấy key, tạo bản ghi như vậy cho từng subdomain cho các ingress trong kubernetes. Webhook này sẽ đảm nhiệm việc tự động tạo các bản ghi đó, và tự động xóa đi khi cert-manager tạo TLS/SSL certificate thành công.

## Cài đặt

### Cài đặt cert manager

Cài đặt cert-manager theo tài liệu chính thức: <https://cert-manager.io/docs/installation/kubernetes/>

### Cài đặt webhook

#### Phương án 1

Cài đặt bizflycloud-certmanager-dns-webhook sử dụng helm

**Lưu ý**: Chọn groupname độc nhất, có thể lấy tên công ty của bạn (ví dụ `acme.mycompany.example`).

Chỉnh sửa các thông tin xác thực trong `./deploy/bizflycloud-certmanager-dns-webhook/values.yaml`

**Lưu ý**: groupName và email là bắt buộc, có thể sử dụng **password** HOẶC **appCredential**

Cài đặt sử dụng câu lệnh

```bash
helm install <deploy name> ./deploy/bizflycloud-certmanager-dns-webhook 
```

#### Phương án 2

Cài đặt bizflycloud-certmanager-dns-webhook với manifest.

**Notes**: Webhook được triển khai thành API Service nhằm giúp quản trị viên hạn chế được các truy cập thông qua Kubernetes RBAC.


Cài đặt trong file `./manifest/bundle.yaml`

Chú ý thay đổi groupname `acme.mycompany.com` match ClusterIssuer trong các trường sau:

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

## Demo

Sau khi cài đặt cert-manager và webhook

1. Tạo 2 service để demo:
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

2. Cài đặt nginx-ingress-controller

    Hướng dẫn: <https://engineering.bizflycloud.vn/cai-dat-nginx-ingress-controller-cho-kubernetes/>

    Sau đó, sử dụng dịch vụ DNS của BizflyCloud trỏ bản ghi và subdomain của bạn, địa chỉ IPv4 lấy từ loadbalancer được tạo ra với nginx-ingress

    ![dns](https://raw.githubusercontent.com/lmq1999/123/master/image.png)

3. Tạo ClusterIssuer/Issuer

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

    **Lưu ý** groupName phải match với groupName khi cài đặt webhook ở trên. Email là email để Let's encrypt thông báo cert sắp hết hạn

4. Tạo nginx ingress

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

5. Đợt 1 lúc để Let’s Encrypt cấp certificate

6. Kiểm tra

```bash
quanlm@quanlm-desktop:~$ curl https://echo2.quanlm1999-testz.tk/
echo2
```

Sử dụng câu lệnh `curl -v` để thấy bắt tay TLS/SSL

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