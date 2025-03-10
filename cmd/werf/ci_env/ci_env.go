package ci_env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	TaggingStrategy string
	Verbose         bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "ci-env CI_SYSTEM",
		DisableFlagsInUseLine: true,
		Short:                 "Generate werf environment variables for specified CI system",
		Long: `Generate werf environment variables for specified CI system.

Currently supported only GitLab CI`,
		Example: `  # Load generated werf environment variables on gitlab job runner
  $ source <(werf ci-env gitlab --tagging-strategy tag-or-branch)`,
		RunE: runCIEnv,
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config")
	common.SetupInsecureRepo(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.TaggingStrategy, "tagging-strategy", "", "", "tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified CI_SYSTEM environment variables")
	cmd.Flags().BoolVarP(&CmdData.Verbose, "verbose", "", false, "Generate echo command for each resulted script line")

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
		return err
	}

	switch CmdData.TaggingStrategy {
	case "tag-or-branch":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided tagging-strategy '%s' not supported", CmdData.TaggingStrategy)
	}

	ciSystem := args[0]

	switch ciSystem {
	case "gitlab":
		err := generateGitlabEnvs()
		if err != nil {
			fmt.Println()
			printError(err.Error())
		}
		return err
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided ci system '%s' not supported", ciSystem)
	}
}

func generateGitlabEnvs() error {
	dockerConfigPath := *CommonCmdData.DockerConfig
	if *CommonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	tmp_manager.AutoGCEnabled = false

	dockerConfig, err := tmp_manager.CreateDockerConfigDir(dockerConfigPath)
	if err != nil {
		return fmt.Errorf("unable to create tmp docker config: %s", err)
	}

	if err := docker_registry.Init(docker_registry.Options{AllowInsecureRepo: *CommonCmdData.InsecureRepo}); err != nil {
		return err
	}

	// Init with new docker config dir
	if err := docker.Init(dockerConfig); err != nil {
		return err
	}

	imagesRepo := os.Getenv("CI_REGISTRY_IMAGE")
	var imagesUsername, imagesPassword string
	doLogin := false
	if imagesRepo != "" {
		isGRC, err := docker_registry.IsGCR(imagesRepo)
		if err != nil {
			return err
		}

		if !isGRC && os.Getenv("CI_JOB_TOKEN") != "" {
			imagesUsername = "gitlab-ci-token"
			imagesPassword = os.Getenv("CI_JOB_TOKEN")
			doLogin = true
		}
	}

	if doLogin {
		err := docker.Login(imagesUsername, imagesPassword, imagesRepo)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", imagesRepo, err)
		}
	}

	var ciGitTag, ciGitBranch string

	if os.Getenv("CI_BUILD_TAG") != "" {
		ciGitTag = os.Getenv("CI_BUILD_TAG")
	} else if os.Getenv("CI_COMMIT_TAG") != "" {
		ciGitTag = os.Getenv("CI_COMMIT_TAG")
	} else if os.Getenv("CI_BUILD_REF_NAME") != "" {
		ciGitBranch = os.Getenv("CI_BUILD_REF_NAME")
	} else if os.Getenv("CI_COMMIT_REF_NAME") != "" {
		ciGitBranch = os.Getenv("CI_COMMIT_REF_NAME")
	}

	printHeader("DOCKER CONFIG", false)
	printExportCommand("DOCKER_CONFIG", dockerConfig)

	printHeader("IMAGES REPO", true)
	printExportCommand("WERF_IMAGES_REPO", imagesRepo)

	printHeader("TAGGING", true)
	if ciGitTag != "" {
		printExportCommand("WERF_TAG_GIT_TAG", slug.DockerTag(ciGitTag))
	}
	if ciGitBranch != "" {
		printExportCommand("WERF_TAG_GIT_BRANCH", slug.DockerTag(ciGitBranch))
	}

	printHeader("DEPLOY", true)
	printExportCommand("WERF_ENV", os.Getenv("CI_ENVIRONMENT_SLUG"))

	printExportCommand("WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL", fmt.Sprintf("project.werf.io/gitlab-url=%s", os.Getenv("CI_PROJECT_URL")))

	cleanupConfig, err := getCleanupConfig()
	if err != nil {
		return fmt.Errorf("unable to get cleanup config: %s", err)
	}

	printHeader("IMAGE CLEANUP POLICIES", true)
	printExportCommand("WERF_GIT_TAG_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyLimit))
	printExportCommand("WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyExpiryDays))
	printExportCommand("WERF_GIT_COMMIT_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyLimit))
	printExportCommand("WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyExpiryDays))

	printHeader("OTHER", true)
	printExportCommand("WERF_LOG_COLOR_MODE", "on")
	printExportCommand("WERF_LOG_PROJECT_DIR", "1")
	printExportCommand("WERF_ENABLE_PROCESS_EXTERMINATOR", "1")
	printExportCommand("WERF_LOG_TERMINAL_WIDTH", "100")

	if ciGitTag == "" && ciGitBranch == "" {
		return fmt.Errorf("none of enviroment variables $WERF_TAG_GIT_TAG=$CI_COMMIT_TAG or $WERF_TAG_GIT_BRANCH=$CI_COMMIT_REF_NAME for '%s' strategy are detected", CmdData.TaggingStrategy)
	}

	return nil
}

func printError(errMsg string) {
	if CmdData.Verbose {
		fmt.Println("echo")
		fmt.Printf("echo 'Error: %s'\n", errMsg)
	}

	fmt.Printf("exit 1\n")
	fmt.Println()
}

func printHeader(header string, withNewLine bool) {
	header = fmt.Sprintf("### %s", header)

	if withNewLine {
		fmt.Println()
	}
	fmt.Println(header)

	if CmdData.Verbose {
		if withNewLine {
			fmt.Println("echo")
		}
		echoHeader := fmt.Sprintf("echo '%s'", header)
		fmt.Println(echoHeader)
	}
}

func printExportCommand(key, value string) {
	exportCommand := fmt.Sprintf("export %s=\"%s\"", key, value)
	if value == "" {
		exportCommand = fmt.Sprintf("# %s", exportCommand)
	}

	fmt.Println(exportCommand)

	if CmdData.Verbose {
		echoExportCommand := fmt.Sprintf("echo '%s'", exportCommand)
		fmt.Println(echoExportCommand)
	}
}

type CleanupConfig struct {
	GitTagStrategyLimit         int `yaml:"gitTagStrategyLimit"`
	GitTagStrategyExpiryDays    int `yaml:"gitTagStrategyExpiryDays"`
	GitCommitStrategyLimit      int `yaml:"gitCommitStrategyLimit"`
	GitCommitStrategyExpiryDays int `yaml:"gitCommitStrategyExpiryDays"`
}

func getCleanupConfig() (CleanupConfig, error) {
	configPath := filepath.Join(werf.GetHomeDir(), "config", "cleanup.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CleanupConfig{
			GitTagStrategyLimit:         10,
			GitTagStrategyExpiryDays:    30,
			GitCommitStrategyLimit:      50,
			GitCommitStrategyExpiryDays: 30,
		}, nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return CleanupConfig{}, fmt.Errorf("error reading %s: %s", configPath, err)
	}

	config := CleanupConfig{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return CleanupConfig{}, fmt.Errorf("bad config yaml %s: %s", configPath, err)
	}

	return config, nil
}
