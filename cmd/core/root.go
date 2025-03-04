package core

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/cli/cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/khulnasoft/dazzle/pkg/fancylog"
)

var rootCfg struct {
	Verbose      bool
	ContextDir   string
	BuildkitAddr string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dazzle",
	Short: "Dazzle is a very experimental Docker image builder with independent layers",
	Long: `Dazzle breaks your usual Docker build by separating the layers. The idea is that
this way we can avoid needless cache invalidation.

THIS IS AN EXPERIEMENT. THINGS WILL BREAK. BEWARE.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		formatter := &fancylog.Formatter{}
		log.SetFormatter(formatter)
		log.SetLevel(log.InfoLevel)

		if rootCfg.Verbose {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	},
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().BoolVarP(&rootCfg.Verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&rootCfg.ContextDir, "context", wd, "context path")
	rootCmd.PersistentFlags().StringVar(&rootCfg.BuildkitAddr, "addr", "unix:///run/buildkit/buildkitd.sock", "address of buildkitd")

	rootCmd.AddCommand(checkCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getResolver() remotes.Resolver {
	dockerCfg := config.LoadDefaultConfigFile(os.Stderr)
	return docker.NewResolver(docker.ResolverOptions{
		Authorizer: docker.NewDockerAuthorizer(docker.WithAuthCreds(func(host string) (user, pwd string, err error) {
			if dockerCfg == nil {
				return
			}

			if host == "registry-1.docker.io" {
				host = "https://index.docker.io/v1/"
			}
			ac, err := dockerCfg.GetAuthConfig(host)
			if err != nil {
				return
			}
			if ac.IdentityToken != "" {
				pwd = ac.IdentityToken
			} else {
				user = ac.Username
				pwd = ac.Password
			}
			log.WithField("host", host).Info("authenticating user")
			return
		})),
	})
}
