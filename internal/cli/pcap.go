package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var pcapCmd = &cobra.Command{
	Use:   "pcap [output.pcap]",
	Short: "Record synthetic probe telemetry and network traffic to Wireshark-compatible PCAP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := args[0]
		count, _ := cmd.Flags().GetInt("count")

		writer, err := netprobe.NewPCAPWriter(outputPath)
		if err != nil {
			return fmt.Errorf("initialize pcap: %w", err)
		}
		defer writer.Close()

		fmt.Printf("Capturing %d synthetic probe packets to %s...\n", count, outputPath)
		for i := 0; i < count; i++ {
			pkt := netprobe.BuildSyntheticTCPPacket([4]byte{127, 0, 0, 1}, [4]byte{127, 0, 0, 1}, uint16(50000+i), 80, []byte("PINGFORGE-PCAP-TEST"))
			if err := writer.WritePacket(time.Now(), pkt); err != nil {
				return fmt.Errorf("write packet: %w", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		fmt.Printf("Successfully recorded %d packets to %s (Wireshark compatible)\n", count, outputPath)
		return nil
	},
}

func init() {
	pcapCmd.Flags().IntP("count", "c", 10, "Number of probe packets to capture")
}
