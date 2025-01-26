package gateway

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	catalogapi "go.bytebuilders.dev/catalog/api/catalog/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	flux "github.com/fluxcd/helm-controller/api/v2"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	kubedbscheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	vgapi "voyagermesh.dev/gateway-api/apis/gateway/v1alpha1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(catalogapi.AddToScheme(scheme))
	utilruntime.Must(kubedbscheme.AddToScheme(scheme))
	utilruntime.Must(gwapiv1.Install(scheme))
	utilruntime.Must(vgapi.AddToScheme(scheme))
	utilruntime.Must(flux.AddToScheme(scheme))
}

func NewCmdGateway(f cmdutil.Factory) *cobra.Command {
	opt := newGatewayOpts(f)
	cmd := &cobra.Command{
		Use:               "gateway",
		Short:             "Gateway related info",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infof("========> %s %s %s \n", opt.dbType, opt.name, opt.ns)
			opt.run()
			fmt.Println("Success !!!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&opt.dbType, "db-type", "t", "mongodb", "Database type")
	cmd.Flags().StringVarP(&opt.name, "name", "m", "mg-test", "Database name")
	cmd.Flags().StringVarP(&opt.ns, "namespace", "n", "demo", "Database namespace")
	return cmd
}

type gatewayOpts struct {
	kc               client.Client
	dbType, name, ns string
}

func newGatewayOpts(f cmdutil.Factory) *gatewayOpts {
	config, err := f.ToRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
	kc, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	return &gatewayOpts{kc: kc}
}

func (g *gatewayOpts) run() {
	g.databases()
}

func (g *gatewayOpts) gateways() error {
	var gwList gwapiv1.GatewayList
	err := g.kc.List(context.TODO(), &gwList, client.InNamespace(g.ns))
	if err != nil {
		log.Fatalf("failed to get gateways: %v", err)
	}

	var uns unstructured.Unstructured
	uns.SetGroupVersionKind(gwapiv1.SchemeGroupVersion.WithKind("GatewayList"))
	err = g.kc.List(context.Background(), &uns, client.InNamespace(g.ns))
	if err != nil {
		log.Fatalf("failed to get gateways: %v", err)
	}
}

func (g *gatewayOpts) databases() {
	var db dbapi.MongoDB
	err := g.kc.Get(context.TODO(), types.NamespacedName{
		Namespace: g.ns,
		Name:      g.name,
	}, &db)
	if err != nil {
		log.Fatalf("failed to get db: %v", err)
	}

}
