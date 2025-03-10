{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build stages for images described in the werf.yaml.

The result of build command are built stages pushed into the specified stages storage (or locally   
in the case when --stages-storage=:local).

If one or more IMAGE_NAME parameters specified, werf will build only these images stages from       
werf.yaml

{{ header }} Syntax

```bash
werf build [IMAGE_NAME...] [options]
```

{{ header }} Examples

```bash
  # Build stages of all images from werf.yaml, built stages will be placed locally
  $ werf stages build --stages-storage :local

  # Build stages of image 'backend' from werf.yaml
  $ werf stages build --stages-storage :local backend

  # Build and enable drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf build --stages-storage :local --introspect-error

  # Set --stages-storage default value using $WERF_STAGES_STORAGE param
  $ export WERF_STAGES_STORAGE=:local
  $ werf build
```

{{ header }} Environments

```bash
  $WERF_DEBUG_ANSIBLE_ARGS  Pass specified cli args to ansible ($ANSIBLE_ARGS)
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and push images into the specified      
            stages storage, to pull base images
  -h, --help=false:
            help for build
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-repo=false:
            Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)
      --introspect-before-error=false:
            Introspect failed stage in the clean state, before running all assembly instructions of 
            the stage
      --introspect-error=false:
            Introspect failed stage in the state, right after running failed assembly instruction
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --ssh-key=[]:
            Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see 
            https://werf.io/documentation/reference/toolbox/ssh.html). Option can be specified      
            multiple times to use multiple keys
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is         
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/documentation/reference/stages_and_images.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

