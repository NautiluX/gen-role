# gen-role kubectl

A `kubectl` plugin to generate roles and clusterroles based on kubectl interactions

Say you want to run a script inside your cluster with a ServiceAccount, and want to know which RBAC is necessary for it.
The script should do the following

```
kubectl run -n default curl --image=curlimages/curl --command sleep 30h
kubectl wait --for=condition=ready pod curl
kubectl exec -n default curl -- curl -s http://platform:8080
kubectl delete pod curl -n default
kubectl get deployment -A
```

By running the commands with the gen-role plugin, the necessary RBAC will be cummulated in the files gen-role.yaml and gen-clusterrole.yaml.
All you need to do is bind them to the right ServiceAccount:

```
kubectl gen-role run -n default curl --image=curlimages/curl --command sleep 30h
kubectl gen-role wait --for=condition=ready pod curl
kubectl gen-role exec -n default curl -- curl -s http://platform:8080
kubectl gen-role delete pod curl -n default
kubectl gen-role get deployment -A
```

The the RBAC files will be populated as follows:

```
$ tail -n +1 gen*.yaml
==> gen-cluster-role.yaml <==
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  name: gen-role-generated-clusterrole
  namespace: gen-role-generated-clusterrole
rules:
- apiGroups:
  - apps/v1
  resourceNames:
  - deployments
  verbs:
  - list

==> gen-role.yaml <==
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  name: gen-role-generated-role
  namespace: gen-role-generated-role
rules:
- apiGroups:
  - ""
  resourceNames:
  - pods
  verbs:
  - post
  - get
  - list
  - watch
  - delete
```
