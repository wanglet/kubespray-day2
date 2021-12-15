package controllers

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clusterv1alpha1 "github.com/wanglet/kubespray-day2/api/v1alpha1"
)

func (r *KubesprayJobReconciler) createInventory(ctx context.Context, req ctrl.Request, job *clusterv1alpha1.KubesprayJob) error {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("getting configmap error: %s", err)
		}
	}
	if cm.ObjectMeta.Name != "" {
		r.Log.Info("Configmap already created, skipping", "namespace:", req.Namespace, "name:", req.Name)
		return nil
	}

	inventory, err := r.buildInventory(ctx, job.Spec.Nodes)
	if err != nil {
		return err
	}
	cm.ObjectMeta.Name = req.Name
	cm.ObjectMeta.Namespace = req.Namespace
	cm.Data = map[string]string{"hosts": string(inventory)}
	if err := controllerutil.SetControllerReference(job, &cm, r.Scheme); err != nil {
		return err
	}

	if err = r.Create(ctx, &cm); err != nil {
		return fmt.Errorf("failed to create cm: %s", err)
	}
	msg := fmt.Sprintf("Created configmap %v", cm.Name)
	r.Log.Info(msg)
	r.Recorder.Event(job, corev1.EventTypeNormal, "Created", msg)

	return err
}

func (r *KubesprayJobReconciler) buildInventory(ctx context.Context, limits []clusterv1alpha1.KubesprayNode) ([]byte, error) {
	var nodes corev1.NodeList
	if err := r.List(ctx, &nodes); err != nil {
		return nil, fmt.Errorf("getting nodes error: %s", err)
	}

	// ungrouped := &group{name: "ungrouped", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{}}
	kubeMaster := &group{name: "kube-master", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{}}
	kubeNode := &group{name: "kube-node", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{}}
	etcd := &group{name: "etcd", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{"kube-master": kubeMaster}}
	k8sCluster := &group{name: "k8s-cluster", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{"kube-master": kubeMaster, "kube-node": kubeNode}}
	all := &group{name: "all", Vars: map[string]string{}, Hosts: map[string]*host{}, Children: map[string]*group{"etcd": etcd, "k8s-cluster": k8sCluster}}
	inventory := inventoryData{All: all}

	for _, n := range nodes.Items {
		h := &host{name: n.ObjectMeta.Name, Vars: map[string]string{}}

		for k, v := range n.ObjectMeta.Annotations {
			if strings.HasPrefix(k, "vars.ansible.com") {
				kSlice := strings.Split(k, "/")
				h.Vars[kSlice[1]] = v
			}
		}
		all.Hosts[h.name] = h

		for _, address := range n.Status.Addresses {
			if address.Type == "InternalIP" {
				h.Vars["ansible_host"] = address.Address
			}
		}

		if _, ok := n.ObjectMeta.Labels["node-role.kubernetes.io/master"]; ok {
			kubeMaster.Hosts[h.name] = h
		} else {
			kubeNode.Hosts[h.name] = h
		}
	}

	for _, n := range limits {
		h := &host{name: n.Name, Vars: map[string]string{}}
		all.Hosts[h.name] = h
		h.Vars["ansible_host"] = n.Host
		if n.User != "" {
			h.Vars["ansible_user"] = n.User
		}
		if n.Password != "" {
			h.Vars["ansible_password"] = n.Password
		}
		if n.Role == "master" {
			kubeMaster.Hosts[h.name] = h
		} else {
			kubeNode.Hosts[h.name] = h
		}
	}

	out, _ := yaml.Marshal(inventory)

	return out, nil
}

type inventoryData struct {
	All *group `yaml:"all"`
}

type group struct {
	name     string
	Hosts    map[string]*host  `yaml:"hosts,omitempty"`
	Vars     map[string]string `yaml:"vars,omitempty"`
	Children map[string]*group `yaml:"children,omitempty"`
}
type host struct {
	name string
	Vars map[string]string `yaml:",inline"`
}
