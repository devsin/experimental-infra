package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	GuardLabelKey   = "platform.cyberark.dev/guard"
	GuardLabelValue = "enabled"

	ManagedByLabelKey   = "app.kubernetes.io/managed-by"
	ManagedByLabelValue = "namespaceguard"
)

type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		// namespace deleted
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Skip terminating namespaces
	if ns.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// Skip system namespaces
	if isSystemNamespace(ns.Name) {
		return ctrl.Result{}, nil
	}

	// Only manage namespaces explicitly opted-in
	if ns.Labels[GuardLabelKey] != GuardLabelValue {
		return ctrl.Result{}, nil
	}

	// Ensure guardrails
	if err := r.ensureResourceQuota(ctx, ns.Name); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureLimitRange(ctx, ns.Name); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureNetworkPolicyDefaultDeny(ctx, ns.Name); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureNetworkPolicyAllowDNS(ctx, ns.Name); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func isSystemNamespace(name string) bool {
	switch name {
	case "kube-system", "kube-public", "kube-node-lease", "default",
		"argocd", "istio-system", "istio-ingress", "gmp-system", "gmp-public":
		return true
	default:
		return false
	}
}

func (r *NamespaceReconciler) ensureResourceQuota(ctx context.Context, namespace string) error {
	rqName := "namespaceguard-quota"
	obj := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: rqName, Namespace: namespace}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabelKey] = ManagedByLabelValue
		obj.SetLabels(labels)

		obj.Spec.Hard = corev1.ResourceList{
			// Adjust as you like. Keep simple for demo.
			corev1.ResourceRequestsCPU:    resourceQty("2"),
			corev1.ResourceRequestsMemory: resourceQty("2Gi"),
			corev1.ResourceLimitsCPU:      resourceQty("4"),
			corev1.ResourceLimitsMemory:   resourceQty("4Gi"),
			corev1.ResourcePods:           resourceQty("50"),
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ensure ResourceQuota: %w", err)
	}
	return nil
}

func (r *NamespaceReconciler) ensureLimitRange(ctx context.Context, namespace string) error {
	lrName := "namespaceguard-limits"
	obj := &corev1.LimitRange{ObjectMeta: metav1.ObjectMeta{Name: lrName, Namespace: namespace}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabelKey] = ManagedByLabelValue
		obj.SetLabels(labels)

		obj.Spec.Limits = []corev1.LimitRangeItem{
			{
				Type: corev1.LimitTypeContainer,
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resourceQty("100m"),
					corev1.ResourceMemory: resourceQty("128Mi"),
				},
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resourceQty("500m"),
					corev1.ResourceMemory: resourceQty("512Mi"),
				},
			},
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ensure LimitRange: %w", err)
	}
	return nil
}

func (r *NamespaceReconciler) ensureNetworkPolicyDefaultDeny(ctx context.Context, namespace string) error {
	npName := "namespaceguard-default-deny"
	obj := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: npName, Namespace: namespace}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabelKey] = ManagedByLabelValue
		obj.SetLabels(labels)

		obj.Spec.PodSelector = metav1.LabelSelector{} // all pods
		obj.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeIngress, netv1.PolicyTypeEgress}
		obj.Spec.Ingress = []netv1.NetworkPolicyIngressRule{} // deny all
		obj.Spec.Egress = []netv1.NetworkPolicyEgressRule{}   // deny all
		return nil
	})
	if err != nil {
		return fmt.Errorf("ensure NetworkPolicy default deny: %w", err)
	}
	return nil
}

func (r *NamespaceReconciler) ensureNetworkPolicyAllowDNS(ctx context.Context, namespace string) error {
	npName := "namespaceguard-allow-dns"
	obj := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: npName, Namespace: namespace}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabelKey] = ManagedByLabelValue
		obj.SetLabels(labels)

		obj.Spec.PodSelector = metav1.LabelSelector{} // all pods
		obj.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeEgress}

		// allow egress to kube-dns in kube-system (UDP/TCP 53)
		obj.Spec.Egress = []netv1.NetworkPolicyEgressRule{
			{
				To: []netv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
						},
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"k8s-app": "kube-dns"},
						},
					},
				},
				Ports: []netv1.NetworkPolicyPort{
					{Protocol: protocolPtr(corev1.ProtocolUDP), Port: intstrPtr(53)},
					{Protocol: protocolPtr(corev1.ProtocolTCP), Port: intstrPtr(53)},
				},
			},
		}
		return nil
	})
	if err != nil {
		// some clusters use different DNS labels; you can relax this later
		return fmt.Errorf("ensure NetworkPolicy allow DNS: %w", err)
	}
	return nil
}
