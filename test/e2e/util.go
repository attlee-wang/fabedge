// Copyright 2021 BoCloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"bytes"
	"fmt"
	"math/rand"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/fabedge/fabedge/test/e2e/framework"
)

func ping(pod corev1.Pod, ip string) error {
	timeout := fmt.Sprint(framework.TestContext.PingTimeout)
	_, _, err := execute(pod, []string{"ping", "-w", timeout, "-c", "1", ip})
	return err
}

func execCurl(pod corev1.Pod, url string) (string, string, error) {
	timeout := fmt.Sprint(framework.TestContext.CurlTimeout)
	return execute(pod, []string{"curl", "-s", "-m", timeout, url})
}

func execute(pod corev1.Pod, cmd []string) (string, string, error) {
	cfg, err := framework.LoadConfig()
	if err != nil {
		return "", "", err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", "", err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), err
}

func getName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, rand.Int31n(1000))
}
