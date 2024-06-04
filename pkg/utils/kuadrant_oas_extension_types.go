package utils

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type RouteObject struct {
	Name       *string                        `json:"name,omitempty"`
	Namespace  *string                        `json:"namespace,omitempty"`
	Hostnames  []gatewayapiv1.Hostname        `json:"hostnames,omitempty"`
	ParentRefs []gatewayapiv1.ParentReference `json:"parentRefs,omitempty"`
	Labels     map[string]string              `json:"labels,omitempty"`
}

type KuadrantOASInfoExtension struct {
	Route *RouteObject `json:"route,omitempty"`
}

func NewKuadrantOASInfoExtension(info *openapi3.Info) (*KuadrantOASInfoExtension, error) {
	type KuadrantOASInfoObject struct {
		// Kuadrant extension
		Kuadrant *KuadrantOASInfoExtension `json:"x-kuadrant,omitempty"`
	}

	data, err := info.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var x KuadrantOASInfoObject
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	return x.Kuadrant, nil
}

type KuadrantRateLimitExtension struct {
	When []kuadrantapiv1beta2.WhenCondition `json:"when,omitempty"`

	Counters []kuadrantapiv1beta2.ContextSelector `json:"counters,omitempty"`

	Rates []kuadrantapiv1beta2.Rate `json:"rates,omitempty"`
}

type KuadrantOASPathExtension struct {
	Disable       *bool                         `json:"disable,omitempty"`
	PathMatchType *gatewayapiv1.PathMatchType   `json:"pathMatchType,omitempty"`
	BackendRefs   []gatewayapiv1.HTTPBackendRef `json:"backendRefs,omitempty"`
	RateLimit     *KuadrantRateLimitExtension   `json:"rate_limit,omitempty"`
}

func (k *KuadrantOASPathExtension) IsDisabled() bool {
	// Set default
	return ptr.Deref(k.Disable, false)
}

func (k *KuadrantOASPathExtension) GetPathMatchType() gatewayapiv1.PathMatchType {
	// Set default
	return ptr.Deref(k.PathMatchType, gatewayapiv1.PathMatchExact)
}

func NewKuadrantOASPathExtension(pathItem *openapi3.PathItem) (*KuadrantOASPathExtension, error) {
	type KuadrantOASPathObject struct {
		// Kuadrant extension
		Kuadrant *KuadrantOASPathExtension `json:"x-kuadrant,omitempty"`
	}

	data, err := pathItem.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var x KuadrantOASPathObject
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	kuadrantExtension := ptr.Deref(x.Kuadrant, KuadrantOASPathExtension{})

	return &kuadrantExtension, nil
}

type KuadrantOASOperationExtension KuadrantOASPathExtension

func NewKuadrantOASOperationExtension(operation *openapi3.Operation) (*KuadrantOASOperationExtension, error) {
	type KuadrantOASOperationObject struct {
		// Kuadrant extension
		Kuadrant *KuadrantOASOperationExtension `json:"x-kuadrant,omitempty"`
	}

	data, err := operation.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var x KuadrantOASOperationObject
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	kuadrantExtension := ptr.Deref(x.Kuadrant, KuadrantOASOperationExtension{})

	return &kuadrantExtension, nil
}
