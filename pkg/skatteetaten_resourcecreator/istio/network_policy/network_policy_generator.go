package network_policy

import (
	"sort"

	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/skatteetaten_resourcecreator/istio/authorization_policy"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	KubeNamespace = "kube-system"
	MetricsPort   = 15020
	DNSPort       = 53
)

type Source interface {
	resource.Source
	GetIngress() *skatteetaten_no_v1alpha1.IngressConfig
	GetEgress() *skatteetaten_no_v1alpha1.EgressConfig
	GetAzure() *skatteetaten_no_v1alpha1.AzureConfig
}

func Create(app Source, ast *resource.Ast) {
	ingressConfig := app.GetIngress()
	egressConfig := app.GetEgress()
	np := generateNetworkPolicy(app)

	// Minimum required policies needed for a pod to start
	np.Spec.Ingress = *generateDefaultIngressRules(app)
	np.Spec.Egress = *generateDefaultEgressRules()

	if ingressConfig != nil {
		// Internal ingress
		// Sort to allow fixture testing
		keys := make([]string, 0, len(ingressConfig.Internal))
		for k := range ingressConfig.Internal {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, rule := range keys {
			np.Spec.Ingress = append(np.Spec.Ingress, *generateNetworkPolicyIngressRule(
				app,
				ingressConfig.Internal[rule]))
		}

		// Public ingress
		for _, ingress := range ingressConfig.Public {
			gateway := ingress.Gateway
			if len(gateway) == 0 {
				gateway = authorization_policy.DefaultIngressGateway
			}

			rule := networkingv1.NetworkPolicyIngressRule{}
			appLabel := map[string]string{
				"app":   gateway,
				"istio": "ingressgateway",
			}

			rule.From = generateNetworkPolicyPeer(authorization_policy.IstioNamespace, appLabel)
			rule.Ports = generateNetworkPolicyPorts([]skatteetaten_no_v1alpha1.PortConfig{{Port: uint16(ingress.Port), Protocol: "TCP"}})
			np.Spec.Ingress = append(np.Spec.Ingress, rule)
		}
	}

	if egressConfig != nil {
		// Internal egress
		// Sort to allow fixture testing
		keys := make([]string, 0, len(egressConfig.Internal))
		for k := range egressConfig.Internal {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, rule := range keys {
			np.Spec.Egress = append(
				np.Spec.Egress, *generateNetworkPolicyEgressRule(
					app,
					egressConfig.Internal[rule]))
		}

		//if we have external integrations or we have an azure resource
		if len(egressConfig.External) > 0 || app.GetAzure() != nil {
			np.Spec.Egress = append(np.Spec.Egress, generateNetworkPolicyExternalEgressRule())
			np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
				To: []networkingv1.NetworkPolicyPeer{{
					IPBlock: &networkingv1.IPBlock{
						CIDR: "10.209.0.0/16",
					},
				}},
				Ports: []networkingv1.NetworkPolicyPort{generateNetworkPolicyPort("TCP", 443)},
			})
		}
	}

	ast.AppendOperation(resource.OperationCreateOrUpdate, np)
}

func generateNetworkPolicy(source resource.Source) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: resource.CreateObjectMeta(source),
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"app": source.GetName()},
			},
		},
	}
}

func generateDefaultIngressRules(source resource.Source) *[]networkingv1.NetworkPolicyIngressRule {
	return &[]networkingv1.NetworkPolicyIngressRule{{
		// Allow prometheus scraping on the "merged metrics" port on the istio proxy.
		// Istio proxy collects metrics from the app on the configured metrics port and merges with own metrics.
		From:  generateNetworkPolicyPeer(authorization_policy.IstioNamespace, map[string]string{"app": "prometheus", "component": "server"}),
		Ports: generateNetworkPolicyPorts([]skatteetaten_no_v1alpha1.PortConfig{{Protocol: "TCP", Port: MetricsPort}}),
	}}
}

func generateDefaultEgressRules() *[]networkingv1.NetworkPolicyEgressRule {
	dnsPort := generateNetworkPolicyPorts([]skatteetaten_no_v1alpha1.PortConfig{{Protocol: "UDP", Port: DNSPort}})

	return &[]networkingv1.NetworkPolicyEgressRule{

		{
			// Allow access to kube-dns
			Ports: dnsPort,
			To:    generateNetworkPolicyPeer(KubeNamespace, map[string]string{"k8s-app": "kube-dns"}),
		},
		{
			// Seems like kube-dns isn't enough. And I am not sure why, but some investigation
			// suggests it is only required when starting up sidecar/init-containers in AKS.
			Ports: dnsPort,
		},
		{
			// Istio Proxy needs access to Istio pilot.
			// TODO: Limit on specific ports.
			To: generateNetworkPolicyPeer(authorization_policy.IstioNamespace, map[string]string{"app": "istiod", "istio": "pilot"}),
		},
		{
			// This is needed to reach the cluster's metadata server (169.254.169.254).
			// It's reachable through localhost, so why we need the rule at all is weird.
			To: []networkingv1.NetworkPolicyPeer{{
				IPBlock: &networkingv1.IPBlock{CIDR: "127.0.0.1/32"},
			}},
		},
	}
}

func generateNetworkPolicyIngressRule(source resource.Source, inbound skatteetaten_no_v1alpha1.InternalIngressConfig) *networkingv1.NetworkPolicyIngressRule {
	appLabel := map[string]string{}

	if inbound.Application != "*" && inbound.Application != "" {
		appLabel["app"] = inbound.Application
	}

	return &networkingv1.NetworkPolicyIngressRule{
		Ports: generateNetworkPolicyPorts(inbound.Ports),
		From:  generateNetworkPolicyPeer(inbound.Namespace, appLabel),
	}
}

func generateNetworkPolicyEgressRule(source resource.Source, outbound skatteetaten_no_v1alpha1.InternalEgressConfig) *networkingv1.NetworkPolicyEgressRule {
	appLabel := map[string]string{}
	if outbound.Application != "*" && outbound.Application != "" {
		appLabel["app"] = outbound.Application
	}

	return &networkingv1.NetworkPolicyEgressRule{
		Ports: generateNetworkPolicyPorts(outbound.Ports),
		To:    generateNetworkPolicyPeer(outbound.Namespace, appLabel),
	}
}

func generateNetworkPolicyExternalEgressRule() networkingv1.NetworkPolicyEgressRule {
	// The Calico version on AKS only supports IP based rules for external hosts. (Calico enterprise
	// supports hostname based filtering). Doing IP-based filtering is not a viable solution, so to
	// allow any external traffic we need accept all. However, we can still force use of Network
	// Policies for any internal traffic. For external egress we use Istio ServiceEntry to handle
	// filtering in Istio. Note that egress has to be configured in Azure firewall (NSG) as well.
	return networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{{
			IPBlock: &networkingv1.IPBlock{
				CIDR: "0.0.0.0/0",
				Except: []string{
					"10.0.0.0/8",
					"172.16.0.0/12",
					"192.168.0.0/16",
				},
			},
		}},
	}
}

func generateNetworkPolicyPeer(namespace string, appLabel map[string]string) []networkingv1.NetworkPolicyPeer {
	np := networkingv1.NetworkPolicyPeer{}

	np.NamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"name": namespace},
	}

	if len(appLabel) > 0 {
		np.PodSelector = &metav1.LabelSelector{
			MatchLabels: appLabel,
		}
	}

	return []networkingv1.NetworkPolicyPeer{np}
}

func generateNetworkPolicyPort(protocol string, port uint16) networkingv1.NetworkPolicyPort {
	protocolType := v1.Protocol(protocol)

	return networkingv1.NetworkPolicyPort{
		Protocol: &protocolType,
		Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: int32(port)},
	}
}

func generateNetworkPolicyPorts(portConfig []skatteetaten_no_v1alpha1.PortConfig) []networkingv1.NetworkPolicyPort {
	var ports []networkingv1.NetworkPolicyPort
	for _, port := range portConfig {
		ports = append(ports, generateNetworkPolicyPort(port.Protocol, port.Port))
	}

	return ports
}
