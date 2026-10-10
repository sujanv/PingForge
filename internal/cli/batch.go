package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/engine"
	"github.com/sujanv/pingforge/internal/output"
)

var (
	batchFile      string
	batchCIDR      string
	batchProbeType string
)

var batchCmd = &cobra.Command{
	Use:   "batch [target1 target2 ...]",
	Short: "Execute concurrent diagnostic probes across multiple targets or CIDRs",
	Long: `Run high-concurrency diagnostic probes against targets supplied from
CLI arguments, text files (-f targets.txt), or CIDR IP ranges (--cidr 192.168.1.0/28).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var targets []core.Target

		for _, arg := range args {
			parsed, err := engine.ParseTargetString(arg)
			if err != nil {
				return err
			}
			targets = append(targets, parsed...)
		}

		if batchFile != "" {
			fileTargets, err := engine.ReadTargetsFile(batchFile)
			if err != nil {
				return fmt.Errorf("read target file %s: %w", batchFile, err)
			}
			targets = append(targets, fileTargets...)
		}

		if batchCIDR != "" {
			cidrTargets, err := engine.ParseTargetString(batchCIDR)
			if err != nil {
				return fmt.Errorf("parse CIDR %s: %w", batchCIDR, err)
			}
			targets = append(targets, cidrTargets...)
		}

		if len(targets) == 0 {
			return fmt.Errorf("no targets specified; provide positional arguments, -f <file>, or --cidr <prefix>")
		}

		probeType := core.ProbeTCP
		switch strings.ToUpper(batchProbeType) {
		case "DNS":
			probeType = core.ProbeDNS
		case "HEALTH":
			probeType = core.ProbeHealth
		case "PING":
			probeType = core.ProbePing
		}

		sp := output.NewSpinner(fmt.Sprintf("Running concurrent %s probes across %d targets...", probeType, len(targets)))
		if !jsonOutput && !yamlOutput {
			sp.Start()
		}

		runner := engine.NewRunner(*globalConfig)
		ctx := context.Background()
		results := runner.ExecuteProbes(ctx, targets, probeType)
		sp.Stop()

		if globalStore != nil {
			_ = globalStore.SaveBatch(context.Background(), results)
		}

		return formatOutput(results)
	},
}

func init() {
	batchCmd.Flags().StringVarP(&batchFile, "file", "f", "", "path to file containing list of targets (one per line)")
	batchCmd.Flags().StringVar(&batchCIDR, "cidr", "", "CIDR notation for scanning IP subnet (e.g. 192.168.1.0/28)")
	batchCmd.Flags().StringVar(&batchProbeType, "type", "TCP", "probe category to run: TCP, DNS, HEALTH, PING")
}
