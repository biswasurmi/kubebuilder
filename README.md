```
# 📖 Guestbook Operator

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?logo=kubernetes&logoColor=white)
![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)
![Kubebuilder](https://img.shields.io/badge/Built%20With-Kubebuilder-green)

A Kubernetes **Operator** that automates the lifecycle of a **Guestbook** application using [Kubebuilder](https://book.kubebuilder.io/).  

It includes:
- 🛠 **Custom Resource Definition (CRD)** for `Guestbook`
- ⚙ **Controller** to reconcile resources
- 🪝 **Webhooks** for defaulting & validation
- ✅ **Tests** for webhook reliability

---

## 🚀 Features

- **Custom Resource**: Define your guestbook app declaratively with YAML.
- **Controller Logic**: Automatically ensures actual cluster state matches desired spec.
- **Webhooks**:
  - Default missing values (`size`, `port`, `auth`).
  - Validate constraints before resources are stored.
- **Tests**: Ginkgo + `envtest` for safe, local simulation.

---

## 🏗 How Kubebuilder Works (Architecture Deep Dive)

Kubebuilder is a **framework** for building Kubernetes APIs and controllers in Go.  
It sets up **all the scaffolding** needed to:
1. Define new resource types (CRDs).
2. Write the logic (controller) that reacts when those resources change.
3. Optionally intercept API requests with **webhooks**.

Think of it as **a factory for Kubernetes operators**.

---

### 1. High-Level Flow

```

+------------------+         +-----------------+         +-----------------+
\|    User (kubectl) | -----> |  Kubernetes API | ----->  | Guestbook CRD   |
+------------------+         +-----------------+         +-----------------+
|
v
+-------------------------+
\|   Guestbook Controller  |
+-------------------------+
\|   Watches & Reconciles
v
+--------------------------------------+
\| Kubernetes Resources (Pods, SVCs...) |
+--------------------------------------+

````

---

### 2. Main Building Blocks

#### **Process (`main.go`)**
- Entry point of the operator.
- Creates a **Manager** and registers all controllers, webhooks, and CRDs.
- Can run **one per cluster** or multiple for high availability.

---

#### **Manager**
The **Manager** is like the conductor of an orchestra — it coordinates everything:
- **Leader Election**: In HA setups, decides which instance is active.
- **Cache**: Watches resources and stores them locally to avoid constant API calls.
- **Clients**: Talks to the Kubernetes API.
- **Controller Lifecycle**: Starts controllers, handles retries, and shuts down gracefully.

---

#### **Controller**
- Watches **one resource kind** — here, the `Guestbook` CRD.
- Uses **Predicates** to filter events (so it doesn’t react to every tiny change).
- Passes events to the **Reconciler**.

---

#### **Reconciler**
- The brain of the operator.
- Gets called with the **desired state** (from the CR) and compares it to the **actual state** in the cluster.
- Makes changes until both match.
- In our case:
  - Create/update Deployment, Service, ConfigMap, etc. for the Guestbook app.
  - Ensure the replica count, image, and settings match `spec`.

---

#### **Webhooks (Optional but powerful)**
- Run before Kubernetes stores a resource.
- Two main types:
  1. **Defaulting** — Fill in missing fields with sensible defaults.
  2. **Validating** — Block invalid configurations.

Example for our `Guestbook` CR:
- Default `size: 1`, `port: 8080`, `auth: false`.
- Validate `size >= 1` and `image` is not empty.

---

### 3. Detailed Event Lifecycle

1. **User applies a CR**:
   ```bash
   kubectl apply -f guestbook.yaml
````

2. **API Server stores it** in etcd and notifies watchers.
3. **Manager’s cache** gets the event → Controller sees a change.
4. **Predicates filter** events → Only relevant changes go to Reconciler.
5. **Reconciler compares** desired vs actual → Issues API calls to fix drift.
6. If **webhooks** are enabled, they intercept requests at step 2.

---

### 4. Why Kubebuilder?

* **Scaffolding**: Generates boilerplate so you can focus on business logic.
* **Best Practices**: Follows the controller-runtime pattern used by Kubernetes itself.
* **Testing Support**: Built-in `envtest` integration.
* **Extensibility**: Add webhooks, admission logic, multi-resource controllers easily.

---

## 🪝 Webhook Rules

**Defaulting**:

* `spec.size` → `1` (if <1)
* `spec.port` → `8080` (if missing)
* `spec.auth` → `false` (if missing)

**Validation**:

* `spec.size` ≥ 1
* `spec.image` must not be empty (create/update)

---

## 📦 Prerequisites

* [Go](https://golang.org/dl/) ≥ 1.20
* [Docker](https://www.docker.com/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Kubebuilder](https://book.kubebuilder.io/getting-started.html#installation)
* [setup-envtest](https://book.kubebuilder.io/reference/envtest.html)
* Kubernetes cluster ([kind](https://kind.sigs.k8s.io/), [minikube](https://minikube.sigs.k8s.io/), etc.)
* Container registry account (e.g., Docker Hub)

---

## ⚡ Quick Start

### 1️⃣ Clone

```bash
git clone https://github.com/urmibiswas/guestbook-operator.git
cd guestbook-operator
```

### 2️⃣ Install Go Dependencies

```bash
go mod tidy
```

### 3️⃣ Setup `envtest` for Tests

```bash
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
setup-envtest use 1.28.0
export KUBEBUILDER_ASSETS=$(setup-envtest use 1.28.0 -p path)
```

*(Add `export ...` to your `~/.bashrc` for persistence)*

### 4️⃣ Run Tests

```bash
go test ./internal/webhook/v1/... -v
```

Expected:

```
SUCCESS! -- 6 Passed | 0 Failed
```

### 5️⃣ Build & Push Image

```bash
make docker-build IMG=urmibiswas/guestbook-controller:v0.2.0
make docker-push IMG=urmibiswas/guestbook-controller:v0.2.0
```

### 6️⃣ Deploy Operator

```bash
kubectl apply -f config/manager/manager.yaml
```

### 7️⃣ Apply Sample CR

```bash
kubectl apply -f config/samples/webapp_v1_guestbook.yaml
```

### 8️⃣ Verify

```bash
kubectl get pods -n default -l app=guestbook-sample
kubectl logs -n default <guestbook-pod-name>
kubectl get guestbook guestbook-sample -o yaml
```

### 9️⃣ Access App

```bash
kubectl port-forward pod/guestbook-sample-pod-0 9090:9090
```

Open 👉 [http://localhost:9090](http://localhost:9090)

### 🔟 Clean Up

```bash
kubectl delete -f config/samples/webapp_v1_guestbook.yaml
kubectl delete -f config/manager/manager.yaml
```

---

## 🧑‍💻 Development

**Regenerate CRDs & Webhooks**

```bash
controller-gen crd webhook paths=./api/v1/... \
  output:crd:dir=config/crd/bases \
  output:webhook:dir=config/webhook
```

**Run Locally**

```bash
make run
```

**Add Tests**

```bash
go test ./internal/webhook/v1/... -v
```



