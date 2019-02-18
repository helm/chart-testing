package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/helm/chart-testing/pkg/exec"
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

// DeleteNamespace deletes the specified namespace. If the namespace does not terminate within 90s, pods running in the
// namespace and, eventually, the namespace itself are force-deleted.
func (k Kubectl) DeleteNamespace(namespace string) {
	fmt.Printf("Deleting namespace '%s'...\n", namespace)
	timeoutSec := "120s"
	if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--timeout", timeoutSec); err != nil {
		fmt.Printf("Namespace '%s' did not terminate after %s.\n", namespace, timeoutSec)
	}

	if _, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Namespace '%s' terminated.\n", namespace)
		return
	}

	fmt.Printf("Namespace '%s' did not terminate after %s.\n", namespace, timeoutSec)

	fmt.Println("Force-deleting pods...")
	if err := k.exec.RunProcess("kubectl", "delete", "pods", "--namespace", namespace, "--all", "--force", "--grace-period=0"); err != nil {
		fmt.Println("Error deleting pods:", err)
	}

	time.Sleep(3 * time.Second)

	if err := k.exec.RunProcess("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Removing finalizers from namespace '%s'...\n", namespace)
		// Getting the namespace json to remove the finalizer
		cmdOutput, err := k.exec.RunProcessAndCaptureOutput(
			"kubectl", "get", "namespace", namespace, "--output=json")
		if err != nil {
			fmt.Println("Error getting namespace json:", err)
			return
		}

		namespaceUpdate := map[string]interface{}{}
		err = json.Unmarshal([]byte(cmdOutput), &namespaceUpdate)
		if err != nil {
			fmt.Println("Error in unmarshalling the payload:", err)
			return
		}
		namespaceUpdate["spec"] = nil
		namespaceUpdateByte, err := json.Marshal(namespaceUpdate)
		if err != nil {
			fmt.Println("Error in marshalling the payload:", err)
			return
		}

		// Remove the finalizers by updating the namespace
		cmdProxy, err := k.exec.RunLongRunningProcess("kubectl", "proxy")
		if err != nil {
			fmt.Println("Error creating the kubectl proxy:", err)
			return
		}

		err = cmdProxy.Start()
		defer cmdProxy.Process.Signal(os.Kill)
		if err != nil {
			fmt.Println("Error starting the kubectl proxy:", err)
			return
		}
		k8sURL := fmt.Sprintf("127.0.0.1:8001/api/v1/namespaces/%s/finalize", namespace)
		req, err := http.NewRequest("PUT", k8sURL, bytes.NewReader(namespaceUpdateByte))
		if err != nil {
			fmt.Println("Error creating the request to update the namespace:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error updating the namespace:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Force-deleting namespace '%s'...\n", namespace)
		if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0", "--ignore-not-found=true"); err != nil {
			fmt.Println("Error deleting namespace:", err)
		}
	}

	time.Sleep(3 * time.Second)

	if err := k.exec.RunProcess("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Println("Namespace still exist:", err)
	}
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

		pods, err := k.GetPodsforDeployment(namespace, deployment)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			pod = strings.Trim(pod, "'")
			ready, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "pod", pod, "--namespace", namespace, "--output",
				`jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
			if err != nil {
				return err
			}
			if ready != "True" {
				return errors.New(fmt.Sprintf("Pods '%s' did not reach ready state!", pod))
			}
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
