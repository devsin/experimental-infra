package controller

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func resourceQty(v string) resource.Quantity {
	return resource.MustParse(v)
}

func protoPtr(p string) *string { return &p }

func protocolPtr(p corev1.Protocol) *corev1.Protocol { return &p }

func intstrPtr(i int) *intstr.IntOrString {
	v := intstr.FromInt(i)
	return &v
}
