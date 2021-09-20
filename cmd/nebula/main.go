package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kubemetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	liberator_scheme "github.com/nais/liberator/pkg/scheme"
	"github.com/nais/naiserator/pkg/controllers"
	"github.com/nais/naiserator/pkg/metrics"
	"github.com/nais/naiserator/pkg/naiserator/config"
	"github.com/nais/naiserator/pkg/readonly"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/synchronizer"
)

func main() {
	err := run()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("Nebula-naiserator shutting down")
}

func run() error {
	var err error

	formatter := log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
	log.SetFormatter(&formatter)
	log.SetLevel(log.DebugLevel)

	log.Info("Nebula-naiserator starting up")

	cfg, err := config.New()
	if err != nil {
		return err
	}


	// Register CRDs with controller-tools
	kscheme, err := liberator_scheme.All()
	if err != nil {
		return err
	}

	kconfig, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	kconfig.QPS = float32(cfg.Ratelimit.QPS)
	kconfig.Burst = cfg.Ratelimit.Burst

	metrics.Register(kubemetrics.Registry)
	mgr, err := ctrl.NewManager(kconfig, ctrl.Options{
		SyncPeriod:         &cfg.Informer.FullSyncInterval,
		Scheme:             kscheme,
		MetricsBindAddress: cfg.Bind,
	})
	if err != nil {
		return err
	}

	if cfg.Features.Webhook {
		// Register create/update validation webhooks for liberator_scheme's CRDs
		if err := liberator_scheme.Webhooks(mgr); err != nil {
			return err
		}
	}

	stopCh := StopCh()

	resourceOptions := resource.NewOptions()
	resourceOptions.AccessPolicyNotAllowedCIDRs = cfg.Features.AccessPolicyNotAllowedCIDRs
	resourceOptions.ApiServerIp = cfg.ApiServerIp
	resourceOptions.AzureratorEnabled = cfg.Features.Azurerator
	resourceOptions.ClusterName = cfg.ClusterName
	resourceOptions.DigdiratorEnabled = cfg.Features.Digdirator
	resourceOptions.DigdiratorHosts = cfg.ServiceHosts.Digdirator
	resourceOptions.GatewayMappings = cfg.GatewayMappings
	resourceOptions.GoogleCloudSQLProxyContainerImage = cfg.GoogleCloudSQLProxyContainerImage
	resourceOptions.GoogleProjectId = cfg.GoogleProjectId
	resourceOptions.HostAliases = cfg.HostAliases
	resourceOptions.JwkerEnabled = cfg.Features.Jwker
	resourceOptions.CNRMEnabled = cfg.Features.CNRM
	resourceOptions.KafkaratorEnabled = cfg.Features.Kafkarator
	resourceOptions.NativeSecrets = cfg.Features.NativeSecrets
	resourceOptions.NetworkPolicy = cfg.Features.NetworkPolicy
	resourceOptions.Proxy = cfg.Proxy
	resourceOptions.Securelogs = cfg.Securelogs
	resourceOptions.SecurePodSecurityContext = cfg.Features.SecurePodSecurityContext
	resourceOptions.VaultEnabled = cfg.Features.Vault
	resourceOptions.Vault = cfg.Vault
	resourceOptions.Wonderwall = cfg.Wonderwall
	//TODO: SKATT
	resourceOptions.SkattUsePullSecret=true
	resourceOptions.Istio=true
	resourceOptions.AzureServiceOperatorEnabled=true


	mgrClient := mgr.GetClient()
	simpleClient, err := client.New(kconfig, client.Options{
		Scheme: kscheme,
	})
	if err != nil {
		return err
	}

	if cfg.DryRun {
		mgrClient = readonly.NewClient(mgrClient)
		simpleClient = readonly.NewClient(simpleClient)
	}


	skatteetatenApplicationReconciler := controllers.NewSkatteetatenAppReconciler(synchronizer.Synchronizer{
		Client:          mgrClient,
		Config:          *cfg,
		Kafka:           nil,
		ResourceOptions: resourceOptions,
		RolloutMonitor:  make(map[client.ObjectKey]synchronizer.RolloutMonitor),
		Scheme:          kscheme,
		SimpleClient:    simpleClient,
	})

	if err = skatteetatenApplicationReconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return mgr.Start(stopCh)
}

func StopCh() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

