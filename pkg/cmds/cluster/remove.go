package cluster

import (
	"errors"
	"fmt"
	"sync"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	ace "go.bytebuilders.dev/client"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func newCmdRemove(f *config.Factory) *cobra.Command {
	opts := clustermodel.RemovalOptions{}
	cmd := &cobra.Command{
		Use:               "remove",
		Short:             "Remove a cluster from ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Components.FeatureSets = defaultFeatureSet
			err := removeCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Cluster has been removed already.")
					return nil
				}
				return fmt.Errorf("failed to remove cluster. Reason: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "remove-fluxcd", true, "Specify whether to remove FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "remove-license-server", true, "Specify whether to remove license server or not (default true).")
	return cmd
}

func removeCluster(f *config.Factory, opts clustermodel.RemovalOptions) error {
	fmt.Println("Removing cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	nc, err := c.NewNatsConnection("ace-cli")
	if err != nil {
		return err
	}
	defer nc.Close()

	responseID := xid.New().String()
	wg := sync.WaitGroup{}
	wg.Add(1)
	done := f.Canceller()
	go func() {
		err := printer.PrintNATSJobSteps(&wg, nc, responseID, done)
		if err != nil {
			fmt.Println("Failed to log removal steps. Reason: ", err)
		}
	}()

	err = c.RemoveCluster(opts, responseID)
	if err != nil {
		close(done)
		return err
	}
	wg.Wait()

	return nil
}
