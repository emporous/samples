package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	managerapi "github.com/uor-framework/uor-client-go/api/services/collectionmanager/v1alpha1"
	"github.com/uor-framework/uor-client-go/config"
	"google.golang.org/protobuf/types/known/structpb"
)

// PullOptions describe configuration options that can
// be set using the pull subcommand.
type PullOptions struct {
	*RootOptions
	Source         string
	Output         string
	AttributeQuery string
}

// NewPullCmd creates a new cobra.Command for the pull subcommand.
func NewPullCmd(rootOpts *RootOptions) *cobra.Command {
	o := PullOptions{RootOptions: rootOpts}

	cmd := &cobra.Command{
		Use:           "pull SOCKET-LOCATION SRC",
		Short:         "Pull a UOR collection based on content or attribute address",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Complete(args))
			cobra.CheckErr(o.Run(cmd.Context()))
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "output location for artifacts")
	cmd.Flags().StringVar(&o.AttributeQuery, "attributes", o.AttributeQuery, "attribute query config path")

	return cmd
}

func (o *PullOptions) Complete(args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}
	o.ServerAddress = args[0]
	o.Source = args[1]

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	o.Output = path.Join(cwd, o.Output)
	return nil
}

func (o *PullOptions) Run(ctx context.Context) error {
	client, err := clientSetup(ctx, o.ServerAddress)
	if err != nil {
		return err
	}

	req := managerapi.Retrieve_Request{
		Source:      o.Source,
		Destination: o.Output,
	}

	if o.AttributeQuery != "" {
		query, err := config.ReadAttributeQuery(o.AttributeQuery)
		if err != nil {
			return err
		}

		filter, err := structpb.NewStruct(query.Attributes)
		if err != nil {
			return err
		}

		req.Filter = filter
	}

	resp, err := client.RetrieveContent(ctx, &req)
	if err != nil {
		return err
	}

	if len(resp.Digests) == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No matching collections")
		return nil
	}

	for _, digest := range resp.Digests {
		fmt.Fprintln(o.IOStreams.Out, digest)
	}

	return nil
}
