// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// TODO: [gwapiv1a2-gwapiv1b1]
// This file can be removed once TLSRoute graduates to gwapiv1b1.

func GroupPtrV1Alpha2(group string) *gwapiv1a2.Group {
	gwGroup := gwapiv1a2.Group(group)
	return &gwGroup
}

func KindPtrV1Alpha2(kind string) *gwapiv1a2.Kind {
	gwKind := gwapiv1a2.Kind(kind)
	return &gwKind
}

func NamespacePtrV1Alpha2(namespace string) *gwapiv1a2.Namespace {
	gwNamespace := gwapiv1a2.Namespace(namespace)
	return &gwNamespace
}

func SectionNamePtrV1Alpha2(sectionName string) *gwapiv1a2.SectionName {
	gwSectionName := gwapiv1a2.SectionName(sectionName)
	return &gwSectionName
}

func PortNumPtrV1Alpha2(port int) *gwapiv1a2.PortNumber {
	pn := gwapiv1a2.PortNumber(port)
	return &pn
}

func UpgradeParentReferences(old []gwapiv1a2.ParentReference) []gwapiv1b1.ParentReference {
	newParentReferences := make([]gwapiv1b1.ParentReference, len(old))
	for i, o := range old {
		newParentReferences[i] = UpgradeParentReference(o)
	}
	return newParentReferences
}

// UpgradeParentReference converts gwapiv1a2.ParentReference to gwapiv1b1.ParentReference
func UpgradeParentReference(old gwapiv1a2.ParentReference) gwapiv1b1.ParentReference {
	upgraded := gwapiv1b1.ParentReference{}

	if old.Group != nil {
		upgraded.Group = GroupPtr(string(*old.Group))
	}

	if old.Kind != nil {
		upgraded.Kind = KindPtr(string(*old.Kind))
	}

	if old.Namespace != nil {
		upgraded.Namespace = NamespacePtr(string(*old.Namespace))
	}

	upgraded.Name = old.Name

	if old.SectionName != nil {
		upgraded.SectionName = SectionNamePtr(string(*old.SectionName))
	}

	if old.Port != nil {
		upgraded.Port = PortNumPtr(int32(*old.Port))
	}

	return upgraded
}

func DowngradeParentReference(old gwapiv1b1.ParentReference) gwapiv1a2.ParentReference {
	downgraded := gwapiv1a2.ParentReference{}

	if old.Group != nil {
		downgraded.Group = GroupPtrV1Alpha2(string(*old.Group))
	}

	if old.Kind != nil {
		downgraded.Kind = KindPtrV1Alpha2(string(*old.Kind))
	}

	if old.Namespace != nil {
		downgraded.Namespace = NamespacePtrV1Alpha2(string(*old.Namespace))
	}

	downgraded.Name = old.Name

	if old.SectionName != nil {
		downgraded.SectionName = SectionNamePtrV1Alpha2(string(*old.SectionName))
	}

	if old.Port != nil {
		downgraded.Port = PortNumPtrV1Alpha2(int(*old.Port))
	}

	return downgraded
}

func UpgradeRouteParentStatuses(routeParentStatuses []gwapiv1a2.RouteParentStatus) []gwapiv1b1.RouteParentStatus {
	var res []gwapiv1b1.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, gwapiv1b1.RouteParentStatus{
			ParentRef:      UpgradeParentReference(rps.ParentRef),
			ControllerName: rps.ControllerName,
			Conditions:     rps.Conditions,
		})
	}

	return res
}

func DowngradeRouteParentStatuses(routeParentStatuses []gwapiv1b1.RouteParentStatus) []gwapiv1a2.RouteParentStatus {
	var res []gwapiv1a2.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, gwapiv1a2.RouteParentStatus{
			ParentRef:      DowngradeParentReference(rps.ParentRef),
			ControllerName: rps.ControllerName,
			Conditions:     rps.Conditions,
		})
	}

	return res
}

// UpgradeBackendRef converts gwapiv1a2.BackendRef to gwapiv1b1.BackendRef
func UpgradeBackendRef(old gwapiv1a2.BackendRef) gwapiv1b1.BackendRef {
	upgraded := gwapiv1b1.BackendRef{}

	if old.Group != nil {
		upgraded.Group = GroupPtr(string(*old.Group))
	}

	if old.Kind != nil {
		upgraded.Kind = KindPtr(string(*old.Kind))
	}

	if old.Namespace != nil {
		upgraded.Namespace = NamespacePtr(string(*old.Namespace))
	}

	upgraded.Name = old.Name

	if old.Port != nil {
		upgraded.Port = PortNumPtr(int32(*old.Port))
	}

	return upgraded
}

func DowngradeBackendRef(old gwapiv1b1.BackendRef) gwapiv1a2.BackendRef {
	downgraded := gwapiv1a2.BackendRef{}

	if old.Group != nil {
		downgraded.Group = GroupPtrV1Alpha2(string(*old.Group))
	}

	if old.Kind != nil {
		downgraded.Kind = KindPtrV1Alpha2(string(*old.Kind))
	}

	if old.Namespace != nil {
		downgraded.Namespace = NamespacePtrV1Alpha2(string(*old.Namespace))
	}

	downgraded.Name = old.Name

	if old.Port != nil {
		downgraded.Port = PortNumPtrV1Alpha2(int(*old.Port))
	}

	return downgraded
}

func NamespaceDerefOrAlpha(namespace *gwapiv1a2.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}
	return defaultNamespace
}
