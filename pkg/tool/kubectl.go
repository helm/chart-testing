package tool

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/helm/chart-testing/pkg/exec"
	"github.com/pkg/errors"
)

type Kubectl struct {
	exec exec.ProcessExecutor
}

func NewKubectl() Kubectl {
	return Kubectl{
		exec: exec.ProcessExecutor{},
	}
}

// DeleteNamespace deletes the specified namespace. If the namespace does not terminate within 90s, pods running in the
// namespace and, eventually, the namespace itself are force-deleted.
func (k Kubectl) DeleteNamespace(namespace string) {
	fmt.Printf("Deleting namespace '%s'...\n", namespace)
	timeoutSec := "120s"
	if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--timeout", timeoutSec); err != nil {
		fmt.Printf("Namespace '%s' did not terminate after %s.", namespace, timeoutSec)
	}

	if _, err := k.exec.RunProcessAndCaptureOutput("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Namespace '%s' terminated.\n", namespace)
		return
	}

	fmt.Printf("Namespace '%s' did not terminate after %s.", namespace, timeoutSec)

	fmt.Println("Force-deleting pods...")
	if err := k.exec.RunProcess("kubectl", "delete", "pods", "--namespace", namespace, "--all", "--force", "--grace-period=0"); err != nil {
		fmt.Println("Error deleting pods:", err)
	}

	time.Sleep(3 * time.Second)

	if err := k.exec.RunProcess("kubectl", "get", "namespace", namespace); err != nil {
		fmt.Printf("Force-deleting namespace '%s'...\n", namespace)
		if err := k.exec.RunProcess("kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0"); err != nil {
			fmt.Println("Error deleting namespace:", err)
		}
	}
}

func (k Kubectl) WaitForDeployments(namespace string) error {
	output, err := k.exec.RunProcessAndCaptureOutput(
		"kubectl", "get", "deployments", "--namespace", namespace, "--output", "jsonpath={.items[*].metadata.name}")
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
