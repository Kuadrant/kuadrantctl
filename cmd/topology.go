package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	topologyNS            string
	topologySVGOutputFile string
	topologyDOTOutputFile string
)

func topologyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topology",
		Short: "Export and visualize kuadrant topology",
		Long:  "Export and visualize kuadrant topology",
		RunE:  runTopology,
	}

	cmd.Flags().StringVarP(&topologyNS, "namespace", "n", "kuadrant-system", "Topology's namespace")
	cmd.Flags().StringVarP(&topologySVGOutputFile, "output", "o", "", "SVG image output file")
	cmd.Flags().StringVarP(&topologyDOTOutputFile, "dot", "d", "", "Graphviz DOT output file")
	err := cmd.MarkFlagRequired("output")
	if err != nil {
		panic(err)
	}
	return cmd
}

func runTopology(cmd *cobra.Command, args []string) error {
	if !strings.HasSuffix(topologySVGOutputFile, ".svg") {
		return errors.New("output file must have .svg extension")
	}
	ctx := cmd.Context()
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
	logf.Log.V(1).Info("reading topology", "object", client.ObjectKeyFromObject(topologyConfigMap), "error", err)
	if err != nil {
		return err
	}

	if topologyDOTOutputFile != "" {
		fDot, err := os.Create(topologyDOTOutputFile)
		if err != nil {
			return err
		}
		defer fDot.Close()

		_, err = fDot.WriteString(topologyConfigMap.Data["topology"])
		logf.Log.V(1).Info("write topology in DOT format to file", "file", topologyDOTOutputFile, "error", err)
		if err != nil {
			return err
		}
	}

	g, err := graphviz.New(ctx)
	if err != nil {
		return err
	}

	graph, err := graphviz.ParseBytes([]byte(topologyConfigMap.Data["topology"]))
	logf.Log.V(1).Info("parse DOT graph", "graph empty", graph == nil, "error", err)
	if err != nil {
		return err
	}

	nodeNum, err := graph.NodeNum()
	logf.Log.V(1).Info("info graph", "graph nodenum", nodeNum, "error", err)
	if err != nil {
		return err
	}

	// 1. write encoded PNG data to buffer
	var buf bytes.Buffer
	err = g.Render(ctx, graph, graphviz.SVG, &buf)
	logf.Log.V(1).Info("render graph to SVG", "buf len", buf.Len(), "error", err)
	if err != nil {
		return err
	}

	// write to file
	fSvg, err := os.Create(topologySVGOutputFile)
	if err != nil {
		return err
	}
	defer fSvg.Close()

	_, err = fSvg.Write(buf.Bytes())
	logf.Log.V(1).Info("write topology in SVG format to file", "file", topologySVGOutputFile, "error", err)
	if err != nil {
		return err
	}

	externalCommand := "xdg-open"
	if _, err := exec.LookPath("open"); err == nil {
		externalCommand = "open"
	}

	openCmd := exec.Command(externalCommand, topologySVGOutputFile)
	// pipe the commands output to the applications
	// standard output
	openCmd.Stdout = os.Stdout

	// Run still runs the command and waits for completion
	// but the output is instantly piped to Stdout
	if err := openCmd.Run(); err != nil {
		return err
	}

	return nil
}
