/*
Copyright 2021 wanglet.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/wanglet/kubespray-day2/api/v1alpha1"
)

func (r *KubesprayJobReconciler) generatePodObject(job *clusterv1alpha1.KubesprayJob) *corev1.Pod {
	command := []string{}
	switch job.Spec.Type {
	case clusterv1alpha1.KubesprayJobTypeScale:
		command = []string{"echo", "ansible-playbook", "scale.yaml"}
	case clusterv1alpha1.KubesprayJobTypeRecoverControlPlane:
		command = []string{"echo", "ansible-playbook", "recover-control-plane.yaml"}
	case clusterv1alpha1.KubesprayJobTypeUpgrade:
		command = []string{"echo", "ansible-playbook", "upgrade.yaml"}
	case clusterv1alpha1.KubesprayJobTypeRemoveNode:
		command = []string{"echo", "ansible-playbook", "remove-node.yaml"}
	}

	command = []string{"sleep", "3600"}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:            job.Name,
					ImagePullPolicy: "IfNotPresent",
					Image:           r.KubesprayImage,
					Command:         command,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "inventory",
							MountPath: "/etc/ansible/hosts",
							SubPath:   "hosts",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "inventory",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: job.Name,
							},
						},
					},
				},
			},
		},
	}
	if job.Spec.ExtraVarsConfigmap != "" {
		volume := corev1.Volume{
			Name: "extravars",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: job.Spec.ExtraVarsConfigmap,
					},
				},
			},
		}
		volumeMount := corev1.VolumeMount{
			Name:      "extravars",
			MountPath: "/etc/ansible/extravars.yaml",
			SubPath:   "extravars.yaml",
		}

		pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, volumeMount)
	}

	return pod
}
