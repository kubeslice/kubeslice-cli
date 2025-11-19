# Testing Strategy

This document outlines the testing strategy for `kubeslice-cli`. We use a two-layered approach: **Unit Tests** for logic isolation & **Integration Tests** for real-world verification.

## 1. Test Architecture

| Type | Location | Purpose | Methodology |
| :--- | :--- | :--- | :--- |
| **Unit Tests** | `_test.go` | Verify internal logic, argument parsing & error handling. | Uses **Fakes** (`FakeExecutor`, `FakeFileSystem`) to test in isolation. |
| **Integration Tests** | `_integration_test.go` | Verify interaction with the host system & external binaries. | Uses **Real Binaries** (`kind`, `helm`, `kubectl`, `docker`) and real file system operations. |

---

## 2. Integration Test Scenarios

The following scenarios are covered by the integration suite in `pkg/internal/`. These tests run against a live environment to ensure end-to-end reliability.

### A. Environment Verification (`verify-executables`)
* **Kind Profile:** Verifies that `kind`, `kubectl`, `docker`, & `helm` are correctly detected on the host.
* **Non-Kind Profile:** Verifies that only `kubectl` and `helm` are checked when `kind` is not required.
* **Failure Handling:** Simulates a missing binary to ensure the CLI exits gracefully with the correct error message.

### B. Helm Repo Operations (`helm-repo-add`)
* **Add Repo:** Verifies the CLI can successfully add a public Helm repository using the real `helm` binary.
* **Local Skip:** Verifies that the `UseLocal: true` configuration correctly skips helm operations.
* **Failure Handling:** Verifies that adding a non-existent repository URL triggers a failure.

### C. Cluster Lifecycle (`kind-clusters`)
* **Create Cluster:** Uses the `kind` binary to spin up real Docker containers for the cluster.
* **Idempotency:** Runs creation twice to ensure the CLI correctly **skips** clusters that already exist.
* **Delete Cluster:** Verifies successful deletion of clusters.
* **Kubeconfig:** Verifies the generation of the `kubeconfig.yaml` file on the real file system.
* **Environment:** Verifies the `KUBECONFIG` environment variable is set correctly.

### D. Network Discovery (`get-network-info`)
* **Kind Path:** Verifies retrieval of the Control Plane IP using `docker inspect` (used for Kind clusters).
* **Non-Kind Path:** Verifies retrieval of connection info using `kubectl config view` and `kubectl get nodes` (used for Cloud/DC clusters).
* **Failure Handling:** Verifies the CLI fails correctly when inspecting a non-existent cluster.

### E. Kubernetes Operations (`kubernetes-operation`)
* **Apply Manifest:** Applies a real Kubernetes manifest (ConfigMap) to a running cluster.
* **Get Resources:** Uses `kubectl get` to verify resources were created successfully.
* **Describe Resources:** Uses `kubectl describe` to inspect system resources.
* **Delete Resources:** Deletes a resource and verifies it is removed from the cluster.

---

## 3. How to Run Tests

### Prerequisites
Ensure the following tools are installed and in your `$PATH`:
* `go` (v1.17+)
* `docker` (Daemon must be running)
* `kind`
* `kubectl`
* `helm`

### Commands

**Run Unit Tests Only** (Fast)
```bash
make test-unit
```

**Run Integration Tests Only** (Slower)
```bash
make test-integration
```

**Run All Tests**
```bash
make test
```