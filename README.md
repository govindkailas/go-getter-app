# go-getter-app
go-getter-app is a simple go microservice that exposes two API endpoints to read and write to VAULT

For the sake of clarity and easiness, let's deploy everything on Kubernetes.
If you dont have K8s cluster, get one from [microk8s](https://microk8s.io/#install-microk8s)

_Create a namespace for vault_
```
k create ns vault 
```

_Add the HashiCorp Helm repository_
```
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update
```

_Deploy the vault helm chart_
```
helm install vault hashicorp/vault --set "server.dev.enabled=true" --namespace vault
```
This will bring up the vault in dev mode and it's not recommended for production use. If you would like to do a production-grade deployment, [check this](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-raft)

_Check the status of helm deployment_
```
kubectl get all -n vault
NAME                                        READY   STATUS    RESTARTS   AGE
pod/vault-agent-injector-7f7f68d457-2dtgp   1/1     Running   0          44s
pod/vault-0                                 1/1     Running   0          44s

NAME                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
service/vault-internal             ClusterIP   None             <none>        8200/TCP,8201/TCP   44s
service/vault-agent-injector-svc   ClusterIP   10.152.183.224   <none>        443/TCP             44s
service/vault                      ClusterIP   10.152.183.65    <none>        8200/TCP,8201/TCP   44s

NAME                                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/vault-agent-injector   1/1     1            1           44s

NAME                                              DESIRED   CURRENT   READY   AGE
replicaset.apps/vault-agent-injector-7f7f68d457   1         1         1       44s

NAME                     READY   AGE
statefulset.apps/vault   1/1     44s
```

_Now that we have vault ready, let's write some secrets_ 
```
kubectl -n vault  exec -it vault-0 -- vault kv put secret/apps/config appaname="go-getter-app" password="My_$tR0nG_Pas$!!" 
```

You should see an output similar to this, take note of the `secret path` mentioned! We will need this later
```
===== Secret Path =====
secret/data/apps/config

======= Metadata =======
Key                Value
---                -----
created_time       2023-12-13T00:05:03.193794332Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
```

We must create a policy in vault to allow read/write operations from/to the secret path. We will be using the same secret path from the last result. 
_Exec to the vault pod as there are multi-line commands to run and it may be error-prone if run from outside_
```
kubectl -n vault  exec -it vault-0 -- sh

# After landing into the vault pod run the below,

vault policy write go-app-rw-policy - <<EOH
path "secret/data/apps/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOH
```
You would see `Success! Uploaded policy: go-app-rw-policy`, if the policy writes were successful. You can now `exit` from the pod.

_The next step is to enable `kubernetes auth` in vault_ 
```
kubectl -n vault  exec -it vault-0 -- vault auth enable kubernetes
# now lets list the auths, kubernetes should be there
kubectl -n vault  exec -it vault-0 -- vault auth list
```

Now, we have to allow vault to communicate to K8s cluster using the `/config` endpoint using the token issued by the service account  [More details here](https://developer.hashicorp.com/vault/docs/auth/kubernetes)
```
kubectl -n vault  exec -it vault-0 -- sh
# After landing into the vault pod run the below,
vault write auth/kubernetes/config token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt issuer="kubernetes/serviceaccount"
```
_Note that it's important to mention `issuer="kubernetes/serviceaccount"` otherwise vault will reject the access as it will not know the token issuer. The issuer might be different if you are on a cloud provider k8s._ 

The next step will be creating `namespace` and `service account` for deploying the `go-getter-app` which would talk to vault and fetch the secrets.
```
kubectl create ns go-app
kubectl -n go-app create serviceaccount go-app-vault-auth-sa
```

_Let's bind it all together_ 😀
```
vault write auth/kubernetes/role/go-app-role \
        bound_service_account_names=go-app-vault-auth-sa \
        bound_service_account_namespaces=go-app \
        policies=go-app-rw-policy \
        ttl=72h
```

More to go ..