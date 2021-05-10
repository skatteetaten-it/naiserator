package networkpolicy

import (
	"net/url"

	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/util"

	"github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	prometheusPodSelectorLabelValue        = "prometheus" // Label value denoting the promethues pod-selector
	prometheusNamespace                    = "nais"       // Which namespace Prometheus is installed in
	nginxNamespac                          = "nginx"      // Which namespace Nginx ingress controller runs in
	networkPolicyDefaultEgressAllowIPBlock = "0.0.0.0/0"  // The default IP block CIDR for the default allow network policies per app
)

func Create(app *nais_io_v1alpha1.Application, options resource.Options, operations *resource.Operations) {
	if !options.NetworkPolicy {
		return
	}

	networkPolicy := &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: app.CreateObjectMeta(),
		Spec:       networkPolicySpec(app, options),
	}

	*operations = append(*operations, resource.Operation{Resource: networkPolicy, Operation: resource.OperationCreateOrUpdate})
}

func labelSelector(label string, value string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			label: value,
		},
	}
}

func networkPolicySpec(app *nais_io_v1alpha1.Application, options resource.Options) networkingv1.NetworkPolicySpec {
	return networkingv1.NetworkPolicySpec{
		PodSelector: *labelSelector("app", app.Name),
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Ingress: ingressPolicy(app, options),
		Egress:  egressPolicy(app, options),
	}
}

func networkPolicyPeer(podLabelName, podLabelValue, namespace string) networkingv1.NetworkPolicyPeer {
	return networkingv1.NetworkPolicyPeer{
		NamespaceSelector: labelSelector("name", namespace),
		PodSelector:       labelSelector(podLabelName, podLabelValue),
	}
}

func networkPolicyIngressRule(peer ...networkingv1.NetworkPolicyPeer) networkingv1.NetworkPolicyIngressRule {
	return networkingv1.NetworkPolicyIngressRule{
		From: peer,
	}
}

func networkPolicyEgressRule(peer ...networkingv1.NetworkPolicyPeer) networkingv1.NetworkPolicyEgressRule {
	return networkingv1.NetworkPolicyEgressRule{
		To: peer,
	}
}

func networkPolicyApplicationRules(rules []nais_io_v1.AccessPolicyRule, options resource.Options) (networkPolicy []networkingv1.NetworkPolicyPeer) {
	for _, rule := range rules {

		// non-local access policy rules do not result in network policies
		if !rule.MatchesCluster(options.ClusterName) {
			continue
		}

		networkPolicyPeer := networkingv1.NetworkPolicyPeer{
			PodSelector: labelSelector("app", rule.Application),
		}

		if rule.Application == "*" {
			networkPolicyPeer = networkingv1.NetworkPolicyPeer{PodSelector: &metav1.LabelSelector{}}
		}

		if rule.Namespace != "" {
			networkPolicyPeer.NamespaceSelector = labelSelector("name", rule.Namespace)
		}

		networkPolicy = append(networkPolicy, networkPolicyPeer)
	}

	return
}

func ingressPolicy(app *nais_io_v1alpha1.Application, options resource.Options) []networkingv1.NetworkPolicyIngressRule {
	rules := make([]networkingv1.NetworkPolicyIngressRule, 0)

	rules = append(rules, networkPolicyIngressRule(networkPolicyPeer("app", prometheusPodSelectorLabelValue, prometheusNamespace)))
	rules = append(rules, networkPolicyIngressRule(networkingv1.NetworkPolicyPeer{
		NamespaceSelector: labelSelector("linkerd.io/is-control-plane", "true"),
	}))
	rules = append(rules, networkPolicyIngressRule(networkingv1.NetworkPolicyPeer{
		NamespaceSelector: labelSelector("linkerd.io/extension", "viz"),
		PodSelector:       labelSelector("component", "tap"),
	}))
	rules = append(rules, networkPolicyIngressRule(networkingv1.NetworkPolicyPeer{
		NamespaceSelector: labelSelector("linkerd.io/extension", "viz"),
		PodSelector:       labelSelector("component", "prometheus"),
	}))

	if len(app.Spec.AccessPolicy.Inbound.Rules) > 0 {
		rules = append(rules, networkPolicyIngressRule(networkPolicyApplicationRules(app.Spec.AccessPolicy.Inbound.Rules, options)...))
	}

	if len(app.Spec.Ingresses) > 0 {
		for _, ingress := range app.Spec.Ingresses {
			ur, err := url.Parse(string(ingress))
			if err != nil {
				continue
			}
			gw := util.ResolveIngressClass(ur.Host, options.GatewayMappings)
			if gw == nil {
				continue
			}
			rules = append(rules, networkPolicyIngressRule(networkingv1.NetworkPolicyPeer{
				PodSelector:       labelSelector("app.kubernetes.io/instance", *gw),
				NamespaceSelector: labelSelector("name", nginxNamespac),
			}))
		}
	}

	return rules
}

func egressPolicy(app *nais_io_v1alpha1.Application, options resource.Options) []networkingv1.NetworkPolicyEgressRule {
	defaultRules := defaultAllowEgress(options)

	if len(app.Spec.AccessPolicy.Outbound.Rules) > 0 {
		appRules := networkPolicyEgressRule(networkPolicyApplicationRules(app.Spec.AccessPolicy.Outbound.Rules, options)...)
		defaultRules = append(defaultRules, appRules)
	}

	if app.Spec.LeaderElection && len(options.GoogleProjectId) > 0 {
		apiServerAccessRule := networkPolicyEgressRule(networkingv1.NetworkPolicyPeer{
			IPBlock: &networkingv1.IPBlock{
				CIDR: options.ApiServerIp,
			},
		})
		defaultRules = append(defaultRules, apiServerAccessRule)
	}

	return defaultRules
}

func defaultAllowEgress(options resource.Options) []networkingv1.NetworkPolicyEgressRule {
	peers := make([]networkingv1.NetworkPolicyPeer, 0, 4)

	peers = append(peers, networkingv1.NetworkPolicyPeer{
		NamespaceSelector: labelSelector("linkerd.io/is-control-plane", "true"),
	})

	peers = append(peers, networkingv1.NetworkPolicyPeer{
		PodSelector: labelSelector("k8s-app", "kube-dns"),
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				// select in all namespaces since labels on kube-system is regularly deleted in GCP
			},
		},
	})

	peers = append(peers, networkingv1.NetworkPolicyPeer{
		IPBlock: &networkingv1.IPBlock{
			CIDR:   networkPolicyDefaultEgressAllowIPBlock,
			Except: options.AccessPolicyNotAllowedCIDRs,
		},
	})

	return []networkingv1.NetworkPolicyEgressRule{
		networkPolicyEgressRule(peers...),
	}
}