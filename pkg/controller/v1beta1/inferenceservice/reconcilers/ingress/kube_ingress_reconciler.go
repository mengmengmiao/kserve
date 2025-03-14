/*
Copyright 2021 The KServe Authors.

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

package ingress

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"knative.dev/pkg/network"

	"github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	v1beta1api "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kserve/kserve/pkg/constants"
	"github.com/kserve/kserve/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	knapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// RawIngressReconciler reconciles the kubernetes ingress
type RawIngressReconciler struct {
	client        client.Client
	scheme        *runtime.Scheme
	ingressConfig *v1beta1api.IngressConfig
}

func NewRawIngressReconciler(client client.Client,
	scheme *runtime.Scheme,
	ingressConfig *v1beta1api.IngressConfig) (*RawIngressReconciler, error) {
	return &RawIngressReconciler{
		client:        client,
		scheme:        scheme,
		ingressConfig: ingressConfig,
	}, nil
}

func createRawURL(isvc *v1beta1api.InferenceService,
	ingressConfig *v1beta1api.IngressConfig) (*knapis.URL, error) {
	var err error
	url := &knapis.URL{}
	url.Scheme = ingressConfig.UrlScheme
	url.Host, err = GenerateDomainName(isvc.Name, isvc.ObjectMeta, ingressConfig)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func generateRule(ingressHost string, componentName string, path string, port int32) netv1.IngressRule {
	pathType := netv1.PathTypePrefix
	rule := netv1.IngressRule{
		Host: ingressHost,
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &pathType,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: componentName,
								Port: netv1.ServiceBackendPort{
									Number: port,
								},
							},
						},
					},
				},
			},
		},
	}
	return rule
}

func isMultiPorts(annotations map[string]string) bool {
	if value, ok := annotations[constants.IsMultiPorts]; ok {
		if value == "true" {
			return true
		}
	}
	return false
}

func generateTLS(annotations map[string]string, rules []netv1.IngressRule, ingressConfig *v1beta1api.IngressConfig) []netv1.IngressTLS {
	var tls []netv1.IngressTLS
	var oneTls netv1.IngressTLS
	oneTls.SecretName = *ingressConfig.IngressTlsSecretName
	var port int
	if value, ok := annotations[constants.PortHttpsTransformer]; ok {
		if value != "" {
			ports := strings.Split(value, ",")
			for i := 0; i < len(ports); i++ {
				port, _ = strconv.Atoi(ports[i])
				for j := 0; j < len(rules); j++ {
					for k := 0; k < len(rules[j].HTTP.Paths); k++ {
						if int32(port) == rules[j].HTTP.Paths[k].Backend.Service.Port.Number && strings.Contains(rules[j].HTTP.Paths[k].Backend.Service.Name, "transformer") {
							oneTls.Hosts = append(oneTls.Hosts, rules[j].Host)
							break
						}
					}
				}
			}
			tls = append(tls, oneTls)
			return tls
		}
	}
	if value, ok := annotations[constants.PortHttps]; ok {
		if value != "" {
			ports := strings.Split(value, ",")
			for i := 0; i < len(ports); i++ {
				port, _ = strconv.Atoi(ports[i])
				for j := 0; j < len(rules); j++ {
					for k := 0; k < len(rules[j].HTTP.Paths); k++ {
						if int32(port) == rules[j].HTTP.Paths[k].Backend.Service.Port.Number {
							oneTls.Hosts = append(oneTls.Hosts, rules[j].Host)
							break
						}
					}
				}
			}
			tls = append(tls, oneTls)
			return tls
		}
	}
	return tls
}

func generateMetadata(isvc *v1beta1api.InferenceService,
	componentType constants.InferenceServiceComponent) metav1.ObjectMeta {
	var name string
	switch componentType {
	case constants.Transformer:
		name = constants.DefaultTransformerServiceName(isvc.Name)
	case constants.Explainer:
		name = constants.DefaultExplainerServiceName(isvc.Name)
	case constants.Predictor:
		name = constants.DefaultPredictorServiceName(isvc.Name)
	}
	//get annotations from isvc
	annotations := utils.Filter(isvc.Annotations, func(key string) bool {
		return !utils.Includes(constants.ServiceAnnotationDisallowedList, key)
	})
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: isvc.Namespace,
		Labels: utils.Union(isvc.Labels, map[string]string{
			constants.InferenceServicePodLabelKey: isvc.Name,
			constants.KServiceComponentLabel:      string(componentType),
		}),
		Annotations: annotations,
	}
	return objectMeta
}

// generateIngressHost return the config domain in configmap.IngressDomain
func generateIngressHost(ingressConfig *v1beta1api.IngressConfig,
	isvc *v1beta1api.InferenceService,
	componentType string,
	topLevelFlag bool) (string, error) {

	metadata := generateMetadata(isvc, constants.InferenceServiceComponent(componentType))
	if !topLevelFlag {
		return GenerateDomainName(metadata.Name, isvc.ObjectMeta, ingressConfig)
	} else {
		return GenerateDomainName(isvc.Name, isvc.ObjectMeta, ingressConfig)
	}
}

func generateMultiPortsRules(host string, isvc *v1beta1.InferenceService) []netv1.IngressRule {
	var rules []netv1.IngressRule
	if len(isvc.Spec.Predictor.Containers) != 0 && len(isvc.Spec.Predictor.Containers[0].Ports) != 0 {
		for _, port := range isvc.Spec.Predictor.Containers[0].Ports {
			rules = append(rules, generateRule(port.Name+"."+host, constants.DefaultPredictorServiceName(isvc.Name), "/", port.ContainerPort))
		}
	} else if isvc.Spec.Predictor.Model != nil && len(isvc.Spec.Predictor.Model.Ports) != 0 {
		for _, port := range isvc.Spec.Predictor.Model.Ports {
			rules = append(rules, generateRule(port.Name+"."+host, constants.DefaultPredictorServiceName(isvc.Name), "/", port.ContainerPort))
		}
	}
	return rules
}

func createRawIngress(scheme *runtime.Scheme, isvc *v1beta1api.InferenceService,
	ingressConfig *v1beta1api.IngressConfig) (*netv1.Ingress, error) {
	if !isvc.Status.IsConditionReady(v1beta1api.PredictorReady) {
		isvc.Status.SetCondition(v1beta1api.IngressReady, &apis.Condition{
			Type:   v1beta1api.IngressReady,
			Status: corev1.ConditionFalse,
			Reason: "Predictor ingress not created",
		})
		return nil, nil
	}
	var rules []netv1.IngressRule
	if isvc.Spec.Transformer != nil {
		if !isvc.Status.IsConditionReady(v1beta1api.TransformerReady) {
			isvc.Status.SetCondition(v1beta1api.IngressReady, &apis.Condition{
				Type:   v1beta1api.IngressReady,
				Status: corev1.ConditionFalse,
				Reason: "Transformer ingress not created",
			})
			return nil, nil
		}
		host, err := generateIngressHost(ingressConfig, isvc, string(constants.Transformer), true)
		if err != nil {
			return nil, fmt.Errorf("failed creating top level transformer ingress host: %v", err)
		}
		transformerHost, err := generateIngressHost(ingressConfig, isvc, string(constants.Transformer), false)
		if err != nil {
			return nil, fmt.Errorf("failed creating transformer ingress host: %v", err)
		}
		if isvc.Spec.Explainer != nil {
			explainerHost, err := generateIngressHost(ingressConfig, isvc, string(constants.Explainer), false)
			if err != nil {
				return nil, fmt.Errorf("failed creating explainer ingress host: %v", err)
			}
			rules = append(rules, generateRule(explainerHost, constants.DefaultExplainerServiceName(isvc.Name), "/", constants.CommonDefaultHttpPort))
		}
		// :predict routes to the transformer when there are both predictor and transformer
		if len(isvc.Spec.Transformer.Containers[0].Ports) != 0 {
			rules = append(rules, generateRule(host, constants.DefaultTransformerServiceName(isvc.Name), "/", isvc.Spec.Transformer.Containers[0].Ports[0].ContainerPort))
			rules = append(rules, generateRule(transformerHost, constants.DefaultTransformerServiceName(isvc.Name), "/", isvc.Spec.Transformer.Containers[0].Ports[0].ContainerPort))
		} else {
			rules = append(rules, generateRule(host, constants.DefaultTransformerServiceName(isvc.Name), "/", constants.CommonDefaultHttpPort))
			rules = append(rules, generateRule(transformerHost, constants.DefaultTransformerServiceName(isvc.Name), "/", constants.CommonDefaultHttpPort))
		}
	} else if isvc.Spec.Explainer != nil {
		if !isvc.Status.IsConditionReady(v1beta1api.ExplainerReady) {
			isvc.Status.SetCondition(v1beta1api.IngressReady, &apis.Condition{
				Type:   v1beta1api.IngressReady,
				Status: corev1.ConditionFalse,
				Reason: "Explainer ingress not created",
			})
			return nil, nil
		}
		host, err := generateIngressHost(ingressConfig, isvc, string(constants.Explainer), true)
		if err != nil {
			return nil, fmt.Errorf("failed creating top level explainer ingress host: %v", err)
		}
		explainerHost, err := generateIngressHost(ingressConfig, isvc, string(constants.Explainer), false)
		if err != nil {
			return nil, fmt.Errorf("failed creating explainer ingress host: %v", err)
		}
		// :predict routes to the predictor when there is only predictor and explainer
		if len(isvc.Spec.Predictor.Containers) != 0 && len(isvc.Spec.Predictor.Containers[0].Ports) != 0 {
			rules = append(rules, generateRule(host, constants.DefaultPredictorServiceName(isvc.Name), "/", isvc.Spec.Predictor.Containers[0].Ports[0].ContainerPort))
		} else {
			rules = append(rules, generateRule(host, constants.DefaultPredictorServiceName(isvc.Name), "/", constants.CommonDefaultHttpPort))
		}
		rules = append(rules, generateRule(explainerHost, constants.DefaultExplainerServiceName(isvc.Name), "/", constants.CommonDefaultHttpPort))
	} else {
		host, err := generateIngressHost(ingressConfig, isvc, string(constants.Predictor), true)
		if err != nil {
			return nil, fmt.Errorf("failed creating top level predictor ingress host: %v", err)
		}
		if isMultiPorts(isvc.Annotations) {
			rules = append(rules, generateMultiPortsRules(host, isvc)...)
		} else {
			if len(isvc.Spec.Predictor.Containers) != 0 && len(isvc.Spec.Predictor.Containers[0].Ports) != 0 {
				rules = append(rules, generateRule(host, constants.DefaultPredictorServiceName(isvc.Name), "/", isvc.Spec.Predictor.Containers[0].Ports[0].ContainerPort))
			} else {
				port, _ := strconv.Atoi(constants.InferenceServiceDefaultHttpPort)
				rules = append(rules, generateRule(host, constants.DefaultPredictorServiceName(isvc.Name), "/", int32(port)))
			}
		}
	}
	//add predictor rule
	predictorHost, err := generateIngressHost(ingressConfig, isvc, string(constants.Predictor), false)
	if err != nil {
		return nil, fmt.Errorf("failed creating predictor ingress host: %v", err)
	}
	if isMultiPorts(isvc.Annotations) {
		rules = append(rules, generateMultiPortsRules(predictorHost, isvc)...)
	} else {
		if len(isvc.Spec.Predictor.Containers) != 0 && len(isvc.Spec.Predictor.Containers[0].Ports) != 0 {
			rules = append(rules, generateRule(predictorHost, constants.DefaultPredictorServiceName(isvc.Name), "/", isvc.Spec.Predictor.Containers[0].Ports[0].ContainerPort))
		} else {
			port, _ := strconv.Atoi(constants.InferenceServiceDefaultHttpPort)
			rules = append(rules, generateRule(predictorHost, constants.DefaultPredictorServiceName(isvc.Name), "/", int32(port)))
		}
	}
	tls := generateTLS(isvc.Annotations, rules, ingressConfig)
	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        isvc.ObjectMeta.Name,
			Namespace:   isvc.ObjectMeta.Namespace,
			Annotations: isvc.Annotations,
		},
		Spec: netv1.IngressSpec{
			IngressClassName: ingressConfig.IngressClassName,
			Rules:            rules,
			TLS:              tls,
		},
	}
	if err := controllerutil.SetControllerReference(isvc, ingress, scheme); err != nil {
		return nil, err
	}
	return ingress, nil
}

func semanticIngressEquals(desired, existing *netv1.Ingress) bool {
	return equality.Semantic.DeepEqual(desired.Spec, existing.Spec)
}

func (r *RawIngressReconciler) Reconcile(isvc *v1beta1api.InferenceService) error {
	ingress, err := createRawIngress(r.scheme, isvc, r.ingressConfig)
	if ingress == nil {
		return nil
	}
	if err != nil {
		return err
	}
	//reconcile ingress
	existingIngress := &netv1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: isvc.Namespace,
		Name:      isvc.Name,
	}, existingIngress)
	if err != nil {
		if apierr.IsNotFound(err) {
			err = r.client.Create(context.TODO(), ingress)
			log.Info("creating ingress", "ingressName", isvc.Name, "err", err)
		} else {
			return err
		}
	} else {
		if !semanticIngressEquals(ingress, existingIngress) {
			err = r.client.Update(context.TODO(), ingress)
			log.Info("updating ingress", "ingressName", isvc.Name, "err", err)
		}
	}
	if err != nil {
		return err
	}
	isvc.Status.URL, err = createRawURL(isvc, r.ingressConfig)
	if err != nil {
		return err
	}
	isvc.Status.Address = &duckv1.Addressable{
		URL: &apis.URL{
			Host:   network.GetServiceHostname(isvc.Name, isvc.Namespace),
			Scheme: r.ingressConfig.UrlScheme,
			Path:   "",
		},
	}
	isvc.Status.SetCondition(v1beta1api.IngressReady, &apis.Condition{
		Type:   v1beta1api.IngressReady,
		Status: corev1.ConditionTrue,
	})
	return nil
}
