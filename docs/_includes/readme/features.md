
- Complete application lifecycle management: build and publish images, deploy application into Kubernetes and cleanup unused images by policies.
- Application build and deploy specification (as many components as needed) completely described in one git repository with source code (single source of truth).
- Build images with Dockerfile or with our syntax to take advantage of incremental rebuilds based on git history and carefully crafted tools.
- Helm 2 compatible chart and complex deploy process with logging, tracking, early errors detection and annotations to customize tracking logic of specific resources.
- Kubernetes clusters scanner and different policies to keep the registry clean.

## Coming soon

- 3-way-merge [#1616](https://github.com/flant/werf/issues/1616).
- Content based tagging [#1184](https://github.com/flant/werf/issues/1184).
- Distributed builds with common registry [#1614](https://github.com/flant/werf/issues/1614).

## Complete features list

### Building

- Conveniently build as many images as needed for a single project.
- Building images by Dockerfile or Stapel builder instructions.
- Parallel builds on a single host (using file locks).
- Distributed builds (coming soon) [#1614](https://github.com/flant/werf/issues/1614).
- Advanced build with Stapel:
  - Incremental rebuilds based on git history.
  - Building image based on another defined one.
  - Building images with Ansible tasks or Shell scripts.
  - Sharing a common cache between builds using mounts.
  - Reducing image size by detaching source data and build tools.
- Debug tools for build process inspection.
- Detailed output.

### Publishing

- Different image tagging strategies:
  - Tag image by git tag, branch or commit.
  - Content based tagging (coming soon) [#1184](https://github.com/flant/werf/issues/1184).
  
### Deploy

- Deploy an application into Kubernetes and check that application is deployed correctly.
  - Track all application resources status.
  - Control of resources readiness.
  - Control of the deployment process with annotations.
- Full visibility both of the deploy process and of the final result.
  - Logging and error reporting.
  - Periodical status reports during deploy process.
  - Easy debugging of problems without unnecessary kubectl invocations.
- Fail CI pipeline fast when problem detected.
  - Early resources failures detection during deploy process without need to wait full timeout.
- Full compatibility with Helm 2.
- Ability to limit deploy user access using RBAC definition (Tiller is compiled into Werf and run from the deploy user outside of cluster).
- Parallel deploys on a single host (using file locks).
- Distributed parallel deploys (coming soon) [#1620](https://github.com/flant/werf/issues/1620).
- Allow continuous delivery of new images tagged by the same name (by git branch for example).

### Cleanup

- Local and Docker registry cleaning by customizable policies.
- Keeping images that used in Kubernetes clusters. Werf scans the following kinds of objects: Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ReplicationController.
