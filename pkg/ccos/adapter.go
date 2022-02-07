package ccos

import (
    "context"
    "errors"
    operatorv1 "github.com/openshift/api/operator/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
    "strings"
)

func IsAdoptMode(ctx context.Context, kubeClient *kubernetes.Clientset) bool {
    if kubeClient == nil {
        klog.V(2).ErrorS(errors.New("invalid kubeClient"), "input nil")
        return false
    }

    pods, err := kubeClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
        LabelSelector: labels.SelectorFromSet(labels.Set{"component": "etcd"}).String(),
    })
    if err != nil {
        // Including not found
        klog.Warningf("finding etcd pod in kube-system namespace with component:etcd label failed or not found, %s", err.Error())
        return false
    }

    klog.V(2).Infof("operator works in adopt mode")
    for _, pod := range pods.Items {
        klog.V(2).Infof("adopt etcd pod: %s", pod.Name)
    }
    return true
}

func InstallerPodMutationFunc(pod *corev1.Pod, nodeName string, operatorSpec *operatorv1.StaticPodOperatorSpec, revision int32) error {
    klog.V(2).Infof("nodeName(%s), managed(%s), revision(%d)", nodeName, string(operatorSpec.ManagementState), revision)

    args := pod.Spec.Containers[0].Args
    for i, arg := range args {
        if strings.HasPrefix(arg, "--pod-manifest-dir=") {
            klog.V(2).Infof("in adopt mode, change installer pod arg --pod-manifest-dir to /etc/kubernetes/fake-manifests")
            args[i] = "--pod-manifest-dir=/etc/kubernetes/fake-manifests"
            break
        }
    }

    return nil
}
