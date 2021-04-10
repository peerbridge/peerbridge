/*
Copyright © 2021 PeerBridge

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var save bool

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage secp256k1 keys",
	Long:  "Manage secp256k1 keys for usage inside the PeerBridge blockchain.",
}

var createKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new secp256k1 keypair",
	Long:  "Create a new secp256k1 keypair for usage inside the PeerBridge blockchain.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		kpair, err := secp256k1.GenerateNewKeyPair("")
		if err != nil {
			return
		}

		msg := fmt.Sprintf(`
Successfully created a new secp256k1 keypair.

Private Key: %s
Public Key: %s`,
			color.Sprintf(kpair.PrivateKey, color.Notice),
			color.Sprintf(kpair.PublicKey, color.Notice),
		)
		fmt.Println(msg)

		if save {
			err = saveConfig(kpair.PrivateKey)
			if err == nil {
				msg := fmt.Sprintf("\nSuccessfully saved config file: %s", color.Sprintf(config, color.Success))
				fmt.Println(msg)
			}
		}

		return
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(createKeyCmd)

	createKeyCmd.Flags().BoolVarP(&save, "save", "s", false, "save generated key to config file (default is $HOME/.peerbridge.yaml)")
}

// Save config file using Viper
//
// If a config file is provided vie the --config
// flag it will be overriden.
//
// If no --config flag is provided and a config file
// exists in the home directory of the user, it will
// overriden with the new config.
//
// If neither config flag is provided, nor a default
// config file exists in the home directory, a new
// config file will be created under the $HOME/.peerbridge.yml
func saveConfig(key string) (err error) {
	// Configure new Viper instance with only host and key flags
	v := viper.New()
	v.Set("host", host)
	v.Set("key", key)

	if config == "" {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		config = filepath.Join(home, ".peerbridge.yml")
	}

	v.SetConfigFile(config)

	err = v.WriteConfig()

	return
}
