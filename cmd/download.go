package cmd

import (
	"fmt"
	"os"

	"github.com/jhelison/go-torrent/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func DownloadCmd() *cobra.Command {
	defaultOutPath := viper.GetString("download.output_path")

	cmd := &cobra.Command{
		Use:   "download [torrent_file] [options]",
		Short: "Download a single torrent file into the output path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check if the file exists
			if file, err := os.Stat(filePath); os.IsNotExist(err) || file.IsDir() {
				return fmt.Errorf("Error: File does not exist: %s\n", filePath)
			}

			// Check if the output path exists
			if file, err := os.Stat(defaultOutPath); os.IsNotExist(err) || !file.IsDir() {
				return fmt.Errorf("Error: Out path does not exist: %s\n", defaultOutPath)
			}

			// Get the torrent object
			torrent, err := client.TorrentFromTorrentFile(filePath)
			if err != nil {
				return err
			}

			// Download the torrent
			err = torrent.Download(defaultOutPath)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// Other flags
	cmd.Flags().StringVar(&defaultOutPath, "output", defaultOutPath, "output path do download")

	return cmd
}
