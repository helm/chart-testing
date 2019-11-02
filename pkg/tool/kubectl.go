package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/helm/chart-testing/v3/pkg/exec"
	"github.com/pkg/errors"
)

type Kubectl struct {
	exec exec.ProcessExecutor
}

func NewKubectl(exec exec.ProcessExecutor) Kubectl {
	return Kubectl{
		exec: exec,
	}
}

// CreateNamespace creates a new namespace with the given name.
func (k Kubectl) CreateNamespace(namespace string) error {
	fmt.Printf("Creating namespace '%s'...\n", namespace)
	return k.exec.RunProcess("kubectl", "create", "namespace", namespace)
}

// DeleteNamespace deletes the specified namespace. If the namespace does not terminate within 120s, pods running in the
// namespace and, eventually, the namespace itself are force-deleted.
func (k Kubectl) DeleteNamespace(namespace string) {
	fmt.Printf("Deleting namespace '%s'...\n", namespace)
	timeoutSec := "180s"
	if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--timeout", timeoutSec); err != nil {
		fmt.Printf("Namespace '%s' did not terminate after %s.\n", namespace, timeoutSec)
	}

	if k.getNamespace(namespace) {
		fmt.Printf("Namespace '%s' did not terminate after %s.\n", namespace, timeoutSec)

		fmt.Println("Force-deleting everything...")
		if err := k.exec.RunProcess("kubectl", "delete", "all", "--namespace", namespace, "--all", "--force", "--grace-period=0"); err != nil {
			fmt.Printf("Error deleting everything in the namespace %v: %v", namespace, err)
		}

		// Give it some more time to be deleted by K8s
		time.Sleep(5 * time.Second)

		if k.getNamespace(namespace) {
			if err := k.forceNamespaceDeletion(namespace); err != nil {
				fmt.Println("Error force deleting namespace:", err)
			}
		}
	}
}

func (k Kubectl) forceNamespaceDeletion(namespace string) error {
	// Getting the namespace json to remove the finalizer
	cmdOutput, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "namespace", namespace, "--output=json")
	if err != nil {
		fmt.Println("Error getting namespace json:", err)
		return err
	}

	namespaceUpdate := map[string]interface{}{}
	err = json.Unmarshal([]byte(cmdOutput), &namespaceUpdate)
	if err != nil {
		fmt.Println("Error in unmarshalling the payload:", err)
		return err
	}
	namespaceUpdate["spec"] = nil
	namespaceUpdateBytes, err := json.Marshal(&namespaceUpdate)
	if err != nil {
		fmt.Println("Error in marshalling the payload:", err)
		return err
	}

	// Remove finalizer from the namespace
	fun := func(port int) error {
		fmt.Printf("Removing finalizers from namespace '%s'...\n", namespace)

		k8sURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/namespaces/%s/finalize", port, namespace)
		req, err := retryablehttp.NewRequest("PUT", k8sURL, bytes.NewReader(namespaceUpdateBytes))
		if err != nil {
			fmt.Println("Error creating the request to update the namespace:", err)
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		errMsg := "Error removing finalizer from namespace"
		client := retryablehttp.NewClient()
		client.Logger = nil
		if resp, err := client.Do(req); err != nil {
			return errors.Wrap(err, errMsg)
		} else if resp.StatusCode != http.StatusOK {
			return errors.New(errMsg)
		}

		return nil
	}

	err = k.exec.RunWithProxy(fun)
	if err != nil {
		return errors.Wrapf(err, "Cannot force-delete namespace '%s'", namespace)
	}

	// Give it some more time to be deleted by K8s
	time.Sleep(5 * time.Second)

	// Check again
	if _, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Namespace '%s' terminated.\n", namespace)
		return nil
	}

	fmt.Printf("Force-deleting namespace '%s'...\n", namespace)
	if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0", "--ignore-not-found=true"); err != nil {
		fmt.Println("Error deleting namespace:", err)
		return err
	}

	return nil
}

func (k Kubectl) WaitForDeployments(namespace string, selector string) error {
	output, err := k.exec.RunProcessAndCaptureOutput(
		"kubectl", "get", "deployments", "--namespace", namespace, "--selector", selector, "--output", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return err
	}

	deployments := strings.Fields(output)
	for _, deployment := range deployments {
		deployment = strings.Trim(deployment, "'")
		err := k.exec.RunProcess("kubectl", "rollout", "status", "deployment", deployment, "--namespace", namespace)
		if err != nil {
			return err
		}

		// 'kubectl rollout status' does not return a non-zero exit code when rollouts fail.
		// We, thus, need to double-check here.
		//
		// Just after rollout, pods from the previous deployment revision may still be in a
		// terminating state.
		unavailable, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "deployment", deployment, "--namespace", namespace, "--output",
			`jsonpath={.status.unavailableReplicas}`)
		if err != nil {
			return err
		}
		if unavailable != "" && unavailable != "0" {
			return fmt.Errorf("%s replicas unavailable", unavailable)
		}
	}

	return nil
}

func (k Kubectl) GetPodsforDeployment(namespace string, deployment string) ([]string, error) {
	jsonString, _ := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "deployment", deployment, "--namespace", namespace, "--output=json")
	var deploymentMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &deploymentMap)
	if err != nil {
		return nil, err
	}

	spec := deploymentMap["spec"].(map[string]interface{})
	selector := spec["selector"].(map[string]interface{})
	matchLabels := selector["matchLabels"].(map[string]interface{})
	var ls string
	for name, value := range matchLabels {
		if ls != "" {
			ls += ","
		}
		ls += fmt.Sprintf("%s=%s", name, value)
	}

	return k.GetPods("--selector", ls, "--namespace", namespace, "--output", "jsonpath={.items[*].metadata.name}")
}

func (k Kubectl) GetPods(args ...string) ([]string, error) {
	kubectlArgs := []string{"get", "pods"}
	kubectlArgs = append(kubectlArgs, args...)
	pods, err := k.exec.RunProcessAndCaptureOutput("kubectl", kubectlArgs)
	if err != nil {
		return nil, err
	}
	return strings.Fields(pods), nil
}

func (k Kubectl) GetEvents(namespace string) error {
	return k.exec.RunProcess("kubectl", "get", "events", "--output", "wide", "--namespace", namespace)
}

func (k Kubectl) DescribePod(namespace string, pod string) error {
	return k.exec.RunProcess("kubectl", "describe", "pod", pod, "--namespace", namespace)
}

func (k Kubectl) Logs(namespace string, pod string, container string) error {
	return k.exec.RunProcess("kubectl", "logs", pod, "--namespace", namespace, "--container", container)
}

func (k Kubectl) GetInitContainers(namespace string, pod string) ([]string, error) {
	return k.GetPods(pod, "--no-headers", "--namespace", namespace, "--output", "jsonpath={.spec.initContainers[*].name}")
}

func (k Kubectl) GetContainers(namespace string, pod string) ([]string, error) {
	return k.GetPods(pod, "--no-headers", "--namespace", namespace, "--output", "jsonpath={.spec.containers[*].name}")
}

func (k Kubectl) getNamespace(namespace string) bool {
	if _, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Namespace '%s' terminated.\n", namespace)
		return false
	}

	return true
}
