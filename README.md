# RAD: Registry AdmissionController Doohickey

RAD is a simple Kubernetes Mutating Admission Controller Webhook that modifies image names with a custom private registry.

By mutating the container image with a customizable registry, identical manifests can be used for deployments to environments that require a custom registry.  Similar to how tools such as [`Istio`](https://istio.io) inject an `Envoy` sidecar using a `MutatingAdmissionController` Webhook, `RAD` alleviates all the messy front end processing of Kubernetes manifests with fire and forget manifest deployments.

## Who this is for

For Kubernetes environments running in airgapped environments, identical manifests can be repeatedly used across multiple environments with various private registry access simply by modifying the MWH prepending parameters.

## Example

Using an existing `Deployment` below:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
```

`RAD` will mutate the image name upon creation/update of the `Pod` to `my.custom.registry/nginx:1.7.9`.  

## Deploying the Webhook

An example deployment of `RAD` that deploys the webhook to the `rad-webhook` namespace is placed in the `deploy` folder of this repository, an overview of the process is below:

1. Create a self signed certificate/key pair and store it as a Kubernetes `Secret`

```bash
./deploy/ssl-cert-gen.sh \
  --service rad \
  --secret rad-certs \
  --namespace rad-webhook
```

2. Patch the `MutatingWebhookConfiguration` with the CA obtained from the target Kubernetes cluster:

```bash
cat ./deploy/mutatingwebhook.yaml.template | deploy/webhook-patch-ca-bundle.sh > deploy/mutatingwebhook.yaml
```

NOTE: Scripts above pulled directly from [here](https://github.com/morvencao/kube-mutating-webhook-tutorial).

3. Deploy the admission controller defined by a `Deployment`, `Service`, and `MutatingWebhookConfiguration`:

```bash
kubectl apply -k deploy/
```

NOTE: Within the `deployment.yaml` is the only important environment variable: `PREPEND_REGISTRY`, that dictates what `RAD` will prepend to container names.  Configure this at will, overlay it with `Kustomize`, `Helm`, or your favorite templating tool.

4. Get on with your life, `RAD` will intercept all `Pod`'s `CREATE` and `UPDATE` hooks and prepend the image name with your desired private registry.

## Running the Tests

Maybe some day...
