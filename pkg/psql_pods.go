package pkg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"strings"
)

var Namespace = Getenv("NAMESPACE", "default")

func GetPostgresPods() ([]v1.Pod, error) {
	clientset, err := ClientSet()

	if err != nil {
		return nil, err
	}

	var postgrespods []v1.Pod

	pods, err := clientset.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{})
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, "postgres") {
				postgrespods = append(postgrespods, pod)
			}
		}
	}

	return postgrespods, nil
}

func CheckPodForLock(pod *v1.Pod) (bool, error) {
	postgresUser := getPostGresUser(pod)

	if postgresUser == "" {
		return false, fmt.Errorf("count not find postgres user for pod %s", pod.Name)
	}

	out, errOut, err := execCommand(pod, "psql -U "+postgresUser+" -c \"select * from databasechangeloglock where lockgranted < current_timestamp - '20 minutes'::interval\"")

	if err != nil {
		return false, err
	}

	if len(errOut) > 0 {
		return false, fmt.Errorf("%s", errOut)
	}

	hasOneRow := strings.Contains(out, "(1 row)")

	return hasOneRow, nil
}

func ResetLiquibaseLock(pod *v1.Pod) error {
	postgresUser := getPostGresUser(pod)

	if postgresUser == "" {
		return fmt.Errorf("count not find postgres user for pod %s", pod.Name)
	}

	_, errOut, err := execCommand(pod, "psql -U "+postgresUser+" -c \"update databasechangeloglock set locked = 'f';\"")

	if err != nil {
		return err
	}

	if len(errOut) > 0 {
		return fmt.Errorf("%s", errOut)
	}

	return err
}

func getPostGresUser(pod *v1.Pod) string {
	postgresUser := ""
	// first we need the postgres user
	if len(pod.Spec.Containers) > 0 {
		container := pod.Spec.Containers[0]
		for _, envVar := range container.Env {
			if envVar.Name == "POSTGRES_USER" {
				postgresUser = envVar.Value
			}
		}
	}
	return postgresUser
}

func execCommand(pod *v1.Pod, command string) (string, string, error) {
	config, err := K8SConfig()
	if err != nil {
		return "", "", err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	scheme := runtime.NewScheme()
	if err = v1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	request := clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command:   []string{"/bin/sh", "-c", command},
			Container: "postgres",
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, parameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return "", "", errors.Wrapf(err, "error creating executor. %s\n", err.Error())
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: buf,
		Stderr: errBuf,
	})

	if err != nil {
		return "", "", errors.Wrapf(err, "Failed executing command %s on %v/%v. %s", command, pod.Namespace, pod.Name, err.Error())
	}

	return buf.String(), errBuf.String(), nil
}
