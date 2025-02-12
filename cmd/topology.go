package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/goccy/go-graphviz"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	topologyNS            string
	topologySVGOutputFile string
	topologyDOTOutputFile string
	watchFlag             bool
)

type configMapWatcher struct {
	client.Client
	updateCh chan struct{}
}

func topologyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topology",
		Short: "Export and visualize Kuadrant topology",
		Long:  "Export and visualize Kuadrant topology, optionally streaming updates",
		RunE:  runTopology,
	}

	cmd.Flags().StringVarP(&topologyNS, "namespace", "n", "kuadrant-system", "Namespace of the topology ConfigMap")
	cmd.Flags().StringVarP(&topologySVGOutputFile, "svg", "s", "", "SVG image output file")
	cmd.Flags().StringVarP(&topologyDOTOutputFile, "dot", "d", "", "Graphviz DOT output file")
	cmd.Flags().BoolVar(&watchFlag, "watch", false, "Enable resource watching for continuous updates")
	return cmd
}

func runTopology(cmd *cobra.Command, args []string) error {
	if topologySVGOutputFile == "" && topologyDOTOutputFile == "" {
		return errors.New("at least one of --svg or --dot must be provided")
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	k8sClient, err := client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	topologyKey := client.ObjectKey{Name: "topology", Namespace: topologyNS}
	topologyConfigMap := &corev1.ConfigMap{}
	err = k8sClient.Get(ctx, topologyKey, topologyConfigMap)
	logf.Log.V(1).Info("Reading topology ConfigMap", "object", topologyKey, "error", err)
	if err != nil {
		return err
	}

	updateCh := make(chan struct{}, 1)
	noOpLogger := zap.New(zap.WriteTo(io.Discard))

	var mgr manager.Manager
	if watchFlag {
		mgr, err = manager.New(configuration, manager.Options{
			Scheme: scheme.Scheme,
			Logger: noOpLogger, // neuter controller-runtime logging
			Metrics: server.Options{
				BindAddress: "0", // disable the metrics server
			},
		})
		if err != nil {
			return err
		}

		watcher := &configMapWatcher{
			Client:   k8sClient,
			updateCh: updateCh,
		}

		pred := predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return isTopologyConfigMap(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isTopologyConfigMap(e.ObjectNew)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return isTopologyConfigMap(e.Object)
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return isTopologyConfigMap(e.Object)
			},
		}

		if err := ctrl.NewControllerManagedBy(mgr).
			For(&corev1.ConfigMap{}).
			WithEventFilter(pred).
			Complete(watcher); err != nil {
			logf.Log.Error(err, "Failed to create ConfigMap watcher")
			return err
		}

		go func() {
			if err := mgr.Start(ctx); err != nil {
				logf.Log.Error(err, "Failed to start manager")
			}
		}()
	}

	if topologyDOTOutputFile != "" {
		if err := writeDOTFile(topologyDOTOutputFile, topologyConfigMap.Data["topology"]); err != nil {
			return err
		}
	}

	if topologySVGOutputFile != "" {
		if err := renderAndWriteSVG(ctx, topologyConfigMap.Data["topology"], topologySVGOutputFile); err != nil {
			return err
		}

		if err := openSVG(topologySVGOutputFile); err != nil {
			logf.Log.Error(err, "Failed to open SVG file")
		}
	}

	if watchFlag {
		go func() {
			for {
				select {
				case <-updateCh:
					logf.Log.Info("Received update signal. Re-rendering outputs.")
					updatedConfigMap := &corev1.ConfigMap{}
					err := k8sClient.Get(ctx, topologyKey, updatedConfigMap)
					if err != nil {
						logf.Log.Error(err, "Failed to re-fetch topology ConfigMap during update")
						continue
					}

					if topologySVGOutputFile != "" {
						if err := renderAndWriteSVG(ctx, updatedConfigMap.Data["topology"], topologySVGOutputFile); err != nil {
							logf.Log.Error(err, "Failed to re-render SVG during update")
							continue
						}
					}

					if topologyDOTOutputFile != "" {
						if err := writeDOTFile(topologyDOTOutputFile, updatedConfigMap.Data["topology"]); err != nil {
							logf.Log.Error(err, "Failed to update DOT file during update")
						}
					}

					logf.Log.Info("Successfully re-rendered outputs and notified clients.")
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	if watchFlag {
		// wait for stop signal only with --watch
		<-stop
		logf.Log.Info("Shutting down gracefully")
	} else {
		logf.Log.Info("Topology export complete, exiting")
	}

	return nil
}

func isTopologyConfigMap(obj client.Object) bool {
	return obj.GetName() == "topology" && obj.GetNamespace() == topologyNS
}

func writeDOTFile(filePath, topologyData string) error {
	fDot, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fDot.Close()

	_, err = fDot.WriteString(topologyData)
	logf.Log.V(1).Info("Wrote topology in DOT format to file", "file", filePath, "error", err)
	if err != nil {
		return err
	}
	return nil
}

func renderAndWriteSVG(ctx context.Context, topologyData, outputFile string) error {
	g, err := graphviz.New(ctx)
	if err != nil {
		return err
	}
	defer g.Close()

	graph, err := graphviz.ParseBytes([]byte(topologyData))
	logf.Log.V(1).Info("Parsed DOT graph", "graph is nil", graph == nil, "error", err)
	if err != nil {
		return err
	}

	nodeNum, err := graph.NodeNum()
	logf.Log.V(1).Info("Graph info", "number of nodes", nodeNum, "error", err)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = g.Render(ctx, graph, graphviz.SVG, &buf)
	logf.Log.V(1).Info("Rendered graph to SVG", "buffer length", buf.Len(), "error", err)
	if err != nil {
		return err
	}

	fSvg, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fSvg.Close()

	_, err = fSvg.Write(buf.Bytes())
	logf.Log.V(1).Info("Wrote topology in SVG format to file", "file", outputFile, "error", err)
	if err != nil {
		return err
	}

	return nil
}

func openSVG(filePath string) error {
	externalCommand := "xdg-open"
	if _, err := exec.LookPath("open"); err == nil {
		externalCommand = "open"
	}

	openCmd := exec.Command(externalCommand, filePath)
	openCmd.Stdout = os.Stdout
	openCmd.Stderr = os.Stderr

	return openCmd.Start()
}

func (w *configMapWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	configMap := &corev1.ConfigMap{}
	err := w.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// The ConfigMap was deleted
			return ctrl.Result{}, nil
		}
		logf.Log.Error(err, "Failed to get ConfigMap", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	if configMap.Name != "topology" || configMap.Namespace != topologyNS {
		panic(fmt.Sprintf("unexpected ConfigMap reconciled: got %s/%s, want %s/%s",
			configMap.Namespace, configMap.Name, topologyNS, "topology"))
	}

	if _, exists := configMap.Data["topology"]; !exists {
		logf.Log.Error(errors.New("topology data not found in ConfigMap"), "ConfigMap missing 'topology' key")
		return ctrl.Result{}, nil
	}

	select {
	case w.updateCh <- struct{}{}:
		logf.Log.V(1).Info("Sent update signal to main function")
	default:
		logf.Log.V(1).Info("Update signal already in queue")
	}

	return ctrl.Result{}, nil
}
